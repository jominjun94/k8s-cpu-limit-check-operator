package controller

import (
	"context"
	"math"
	"time"

	reaperv1alpha1 "github.com/jominjun94/k8s-cpu-limit-check-operator/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type CpuReaperPolicyReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	hotPods map[types.UID]time.Time
}

func (r *CpuReaperPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if r.hotPods == nil {
		r.hotPods = make(map[types.UID]time.Time)
	}

	var policy reaperv1alpha1.CpuReaperPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	selector, err := metav1.LabelSelectorAsSelector(policy.Spec.PodSelector)
	if err != nil {
		return ctrl.Result{}, err
	}

	var pods corev1.PodList
	if err := r.List(ctx, &pods, client.InNamespace(policy.Namespace), client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("Pod list", "count", len(pods.Items))

	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			delete(r.hotPods, pod.UID)
			continue
		}

		var metrics metricsv1beta1.PodMetrics
		if err := r.Get(ctx, types.NamespacedName{
			Name: pod.Name, Namespace: pod.Namespace,
		}, &metrics); err != nil {
			logger.Info("Failed to get pod metrics", "pod", pod.Name, "error", err)
			continue
		}

		used := cpuUsageMilli(&metrics)
		limit, ok := cpuLimitMilli(&pod)
		if !ok || limit == 0 {
			continue
		}

		percent := int(math.Round(float64(used) * 100 / float64(limit)))

		logger.Info("CPU check",
			"pod", pod.Name,
			"usedMilli", used,
			"limitMilli", limit,
			"percent", percent,
			"threshold", policy.Spec.ThresholdPercent,
		)

		if percent >= policy.Spec.ThresholdPercent {
			start, exists := r.hotPods[pod.UID]
			if !exists {
				r.hotPods[pod.UID] = time.Now()
				continue
			}

			if time.Since(start) >= time.Duration(policy.Spec.ForSeconds)*time.Second {
				logger.Info("CPU limit exceeded → deleting pod", "pod", pod.Name)
				_ = r.Delete(ctx, &pod)
				delete(r.hotPods, pod.UID)
			}
		} else {
			delete(r.hotPods, pod.UID)
		}
	}

	return ctrl.Result{
		RequeueAfter: time.Duration(policy.Spec.CheckIntervalSeconds) * time.Second,
	}, nil
}

func cpuUsageMilli(m *metricsv1beta1.PodMetrics) int64 {
	var total int64
	for _, c := range m.Containers {
		total += c.Usage.Cpu().MilliValue()
	}
	return total
}

func cpuLimitMilli(p *corev1.Pod) (int64, bool) {
	var total int64
	for _, c := range p.Spec.Containers {
		if limit, ok := c.Resources.Limits[corev1.ResourceCPU]; ok {
			total += limit.MilliValue()
		} else if req, ok := c.Resources.Requests[corev1.ResourceCPU]; ok {
			total += req.MilliValue()
		}
	}
	if total == 0 {
		return 0, false
	}
	return total, true
}

func (r *CpuReaperPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&reaperv1alpha1.CpuReaperPolicy{}).
		Complete(r)
}
