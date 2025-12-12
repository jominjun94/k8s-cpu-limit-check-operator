# CPU Reaper Operator

CPU Reaper Operator는 Kubernetes 클러스터에서 실행 중인 Pod의 CPU 사용률을 metrics.k8s.io API를 통해 주기적으로 확인하고,
설정된 임계치를 초과한 Pod를 자동으로 삭제하여 Deployment, ReplicaSet 등에 의해 Pod가 재생성되도록 유도하는 Operator입니다.

CPU limit은 컨테이너를 종료하지 않고 throttling만 수행하기 때문에
CPU 과다 사용 상태에서도 Pod는 Running 상태로 유지될 수 있으며,
CPU Reaper Operator는 이러한 한계를 보완하기 위해 설계되었습니다.


---

## 주요 기능

- Pod CPU 사용률 모니터링
- CPU Limit(또는 Request)을 기준으로 사용률 계산
- 일정 시간 이상 임계치 초과 시 Pod 자동 삭제
- Deployment / ReplicaSet 환경에서 자동 복구 유도
- 정책(Custom Resource) 기반 설정

---

## 동작 개요

1. 사용자가 `CpuReaperPolicy` 리소스를 생성합니다.
2. Controller는 주기적으로 정책을 Reconcile 합니다.
3. 지정된 Label Selector에 해당하는 Pod를 조회합니다.
4. `metrics.k8s.io` API를 통해 Pod의 CPU 사용량을 조회합니다.
5. Pod의 CPU Limit(없을 경우 Request) 대비 사용률을 계산합니다.
6. 사용률이 임계치를 초과한 상태가 일정 시간 이상 지속되면 Pod를 삭제합니다.
7. 상위 컨트롤러에 의해 Pod가 자동으로 재생성됩니다.

---

## 사전 요구 사항

- Kubernetes 클러스터
- metrics-server 설치 필수
---

## 설치 방법 
```
kubectl apply -f https://raw.githubusercontent.com/jominjun94/k8s-cpu-limit-check-operator/main/dist/install.yaml
```
---

## 정책 생성 (CpuReaperPolicy)
```
apiVersion: reaper.cpu.limit.check/v1alpha1
kind: CpuReaperPolicy
metadata:
  name: cpu-reaper
  namespace: default
spec:
  podSelector:
    matchLabels:
      app: stress
  thresholdPercent: 100
  forSeconds: 30
  checkIntervalSeconds: 10
```
---
## 적용
```
kubectl apply -f cpureaperpolicy.yaml
```
```
테스트용 CPU 부하 Pod
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cpu-stress
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: stress
  template:
    metadata:
      labels:
        app: stress
    spec:
      containers:
      - name: stress
        image: busybox
        command:
          - sh
          - -c
          - |
            while true; do :; done
        resources:
          requests:
            cpu: "100m"
          limits:
            cpu: "100m"
```
---
## 동작 확인
```
kubectl logs -n cpu-reaper-system deploy/cpu-reaper-operator-controller-manager
```
---
## 제거
```
1. kubectl delete cpureaperpolicy cpu-reaper -n default
2. make undeploy
3. make uninstall
```
---
## 컨테이너 이미지
```
jominjun/cpu-reaper-operator:v0.1.0 (public docker registry)
```
