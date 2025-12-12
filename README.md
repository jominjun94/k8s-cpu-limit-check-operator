# CPU Reaper Operator

> Kubernetes Pod CPU 사용률을 실시간으로 감시하고,  
> 설정된 임계치를 초과한 Pod를 자동으로 재기동(삭제)하는 Kubernetes Operator

---

## 📌 프로젝트 개요

**CPU Reaper Operator**는 Kubernetes 클러스터에서 실행 중인 Pod의  
CPU 사용량을 `metrics.k8s.io` 기반으로 주기적으로 확인하여,

- CPU 사용률이 설정한 임계치(%)를 초과하고
- 일정 시간 이상 지속될 경우

해당 Pod를 **자동으로 삭제**하여  
Deployment / ReplicaSet에 의해 **Pod가 재생성되도록 유도**하는  
**정책 기반(Self-Healing) 오퍼레이터**입니다.

운영 환경에서 다음과 같은 상황을 해결하는 것을 목표로 합니다:

- CPU 폭주로 인한 서비스 성능 저하
- 비정상 Pod의 수동 재기동 반복
- HPA만으로 해결하기 어려운 순간적 CPU 스파이크

---

## 🧠 아키텍처 개요

```text
CpuReaperPolicy (Custom Resource)
        │
        ▼
CPU Reaper Controller
        │
        ├─ metrics.k8s.io (PodMetrics)
        │
        ├─ CPU 사용률 계산
        │
        ├─ 임계치 초과 여부 판단
        │
        └─ 임계치 초과 시 Pod 삭제
                    │
                    └─ Deployment / ReplicaSet에 의해 Pod 

⚙️ 동작 방식

사용자가 CpuReaperPolicy CR을 생성

Controller가 주기적으로 정책을 Reconcile

Label Selector에 매칭되는 Pod 목록 조회

metrics.k8s.io API를 통해 Pod CPU 사용량 조회

Pod의 CPU Limit(Request fallback) 대비 사용률 계산

임계치 초과 상태가 forSeconds 이상 지속되면 Pod 삭제

상위 컨트롤러(Deployment/RS)에 의해 Pod 자동 재생성

📦 필수 요구사항

Kubernetes v1.23+

metrics-server 설치 필수
(PodMetrics API 사용)

kubectl get apiservices | grep metrics.k8s.io


Go v1.21+ (개발 시)

Docker / Podman (이미지 빌드 시)

🧩 Custom Resource 정의 (CpuReaperPolicy)
apiVersion: reaper.cpu.limit.check/v1alpha1
kind: CpuReaperPolicy
metadata:
  name: cpu-reaper
  namespace: default
spec:
  podSelector:
    matchLabels:
      app: stress
  thresholdPercent: 100        # CPU 사용률 %
  forSeconds: 30               # 초과 상태 유지 시간
  checkIntervalSeconds: 10     # 체크 주기

🔬 테스트용 CPU 부하 Deployment 예시
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

🚀 설치 방법 (사용자 기준)
1️⃣ CRD 설치
kubectl apply -f https://raw.githubusercontent.com/jominjun94/k8s-cpu-limit-check-operator/main/dist/install.yaml

2️⃣ CpuReaperPolicy 생성
kubectl apply -f cpureaperpolicy.yaml

3️⃣ 동작 확인
kubectl logs -n cpu-reaper-system deploy/cpu-reaper-operator-controller-manager

🐳 컨테이너 이미지
jominjun/cpu-reaper-operator:v0.1.0


Docker Hub 공개 이미지로 별도 인증 없이 Pull 가능

🧪 로컬 개발 모드
make install
make run

📈 향후 개선 예정

HPA 연동

CPU Throttling 기반 판단

Memory 정책 추가

Prometheus / Alertmanager 연계

Dry-Run 모드 지원

👨‍💻 작성자

GitHub: https://github.com/jominjun94

Project: https://github.com/jominjun94/k8s-cpu-limit-check-operator
