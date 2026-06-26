# Kubernetes部署文档

## 概述

本文档描述了如何在Kubernetes集群中部署Tanzanite API服务。

## 架构

```
                    ┌─────────────────┐
                    │   Ingress       │
                    │  (nginx)        │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   Service       │
                    │  (ClusterIP)    │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
   ┌────▼────┐         ┌────▼────┐         ┌────▼────┐
   │  Pod 1  │         │  Pod 2  │         │  Pod 3  │
   │  (API)  │         │  (API)  │         │  (API)  │
   └─────────┘         └─────────┘         └─────────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
        ┌────────────────────┴────────────────────┐
        │                                         │
   ┌────▼──────┐                          ┌──────▼─────┐
   │ PostgreSQL │                          │   Redis    │
   └────────────┘                          └────────────┘
```

## 前置要求

- Kubernetes 1.20+
- kubectl CLI
- Nginx Ingress Controller
- cert-manager（用于HTTPS证书）
- Metrics Server（用于HPA）

## 快速开始

### 1. 创建命名空间（可选）

```bash
kubectl create namespace tanzanite
```

### 2. 配置Secret

编辑 `deployment.yaml` 中的 Secret 部分，设置实际的密钥：

```bash
# 使用kubectl创建secret
kubectl create secret generic tanzanite-secrets \
  --from-literal=jwt-secret='your-secure-jwt-secret' \
  --from-literal=db-username='tanzanite' \
  --from-literal=db-password='your-secure-db-password' \
  --from-literal=redis-password='your-secure-redis-password' \
  -n default
```

### 3. 配置ConfigMap

编辑 `deployment.yaml` 中的 ConfigMap，更新数据库和Redis地址：

```bash
kubectl apply -f k8s/deployment.yaml
```

### 4. 部署应用

```bash
# 部署主应用
kubectl apply -f k8s/deployment.yaml

# 部署HPA（水平自动扩缩）
kubectl apply -f k8s/hpa.yaml

# 部署Ingress
kubectl apply -f k8s/ingress.yaml
```

### 5. 验证部署

```bash
# 查看Pod状态
kubectl get pods -l app=tanzanite-api

# 查看Service
kubectl get svc tanzanite-api

# 查看Ingress
kubectl get ingress tanzanite-api-ingress

# 查看HPA状态
kubectl get hpa tanzanite-api-hpa

# 查看日志
kubectl logs -f deployment/tanzanite-api
```

## 健康检查

应用提供了三个健康检查端点：

### 1. Liveness Probe（存活探针）
- **端点**: `/liveness`
- **用途**: 检测容器是否存活
- **配置**:
  - initialDelaySeconds: 30
  - periodSeconds: 10
  - failureThreshold: 3

### 2. Readiness Probe（就绪探针）
- **端点**: `/readiness`
- **用途**: 检测容器是否准备好接收流量
- **配置**:
  - initialDelaySeconds: 10
  - periodSeconds: 5
  - failureThreshold: 3

### 3. Startup Probe（启动探针）
- **端点**: `/liveness`
- **用途**: 给应用更多时间启动
- **配置**:
  - periodSeconds: 10
  - failureThreshold: 30

### 测试健康检查

```bash
# 端口转发到本地
kubectl port-forward deployment/tanzanite-api 9000:9000

# 测试端点
curl http://localhost:9000/health
curl http://localhost:9000/liveness
curl http://localhost:9000/readiness
```

## 限流配置

### 应用层限流

在代码中配置了多层限流：

1. **全局限流**: 1000 RPS
2. **IP限流**: 100 RPS per IP
3. **用户限流**: 50 RPS per user
4. **认证端点**: 5 RPS
5. **支付端点**: 10 RPS

### Ingress层限流

在 `ingress.yaml` 中配置：

```yaml
annotations:
  nginx.ingress.kubernetes.io/limit-rps: "100"
  nginx.ingress.kubernetes.io/limit-connections: "10"
  nginx.ingress.kubernetes.io/rate-limit: "100"
```

## 自动扩缩

HPA配置了基于CPU和内存的自动扩缩：

- **最小副本数**: 3
- **最大副本数**: 10
- **CPU目标**: 70%
- **内存目标**: 80%

### 扩缩策略

- **扩容**: 快速扩容（30秒内可扩容100%或4个Pod）
- **缩容**: 稳定缩容（5分钟窗口期，最多缩容50%或2个Pod）

### 监控HPA

```bash
# 查看HPA状态
kubectl get hpa tanzanite-api-hpa

# 查看详细信息
kubectl describe hpa tanzanite-api-hpa

# 实时监控
watch kubectl get hpa tanzanite-api-hpa
```

## 滚动更新

### 更新策略

- **maxSurge**: 1（最多多出1个Pod）
- **maxUnavailable**: 0（确保零停机）

### 执行更新

```bash
# 更新镜像
kubectl set image deployment/tanzanite-api api=tanzanite-api:v2

# 或者应用新的配置
kubectl apply -f k8s/deployment.yaml

# 监控更新过程
kubectl rollout status deployment/tanzanite-api

# 查看更新历史
kubectl rollout history deployment/tanzanite-api

# 回滚到上一版本
kubectl rollout undo deployment/tanzanite-api

# 回滚到指定版本
kubectl rollout undo deployment/tanzanite-api --to-revision=2
```

## 资源配置

### Pod资源限制

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### 建议配置

| 环境 | Replicas | CPU Request | Memory Request | CPU Limit | Memory Limit |
|------|----------|-------------|----------------|-----------|--------------|
| 开发 | 1 | 100m | 128Mi | 200m | 256Mi |
| 测试 | 2 | 250m | 256Mi | 500m | 512Mi |
| 生产 | 3-10 | 250m | 256Mi | 500m | 512Mi |

## 日志管理

### 查看日志

```bash
# 查看特定Pod的日志
kubectl logs <pod-name>

# 实时跟踪日志
kubectl logs -f <pod-name>

# 查看前一个容器的日志（如果容器重启了）
kubectl logs <pod-name> --previous

# 查看所有Pod的日志
kubectl logs -l app=tanzanite-api --tail=100

# 使用stern工具（推荐）
stern tanzanite-api
```

### 日志聚合

建议使用以下方案之一进行日志聚合：

1. **EFK Stack** (Elasticsearch + Fluentd + Kibana)
2. **Loki Stack** (Loki + Promtail + Grafana)
3. **云服务** (AWS CloudWatch, GCP Logging, Azure Monitor)

## 监控

### Prometheus指标

应用暴露了 `/metrics` 端点，可以被Prometheus抓取。

### 配置Prometheus

```yaml
apiVersion: v1
kind: ServiceMonitor
metadata:
  name: tanzanite-api-monitor
spec:
  selector:
    matchLabels:
      app: tanzanite-api
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

### Grafana仪表板

推荐导入以下指标：

- HTTP请求率和延迟
- 错误率
- Pod CPU和内存使用
- 数据库连接池状态
- Redis连接状态

## 故障排查

### Pod无法启动

```bash
# 查看Pod状态
kubectl describe pod <pod-name>

# 查看事件
kubectl get events --sort-by='.lastTimestamp'

# 查看日志
kubectl logs <pod-name>
```

### 常见问题

1. **ImagePullBackOff**
   - 检查镜像名称和标签
   - 确认镜像存在于仓库中
   - 检查拉取权限

2. **CrashLoopBackOff**
   - 查看容器日志
   - 检查配置和环境变量
   - 验证数据库和Redis连接

3. **健康检查失败**
   - 检查探针配置
   - 增加initialDelaySeconds
   - 查看应用日志

### 调试命令

```bash
# 进入容器
kubectl exec -it <pod-name> -- /bin/sh

# 端口转发
kubectl port-forward <pod-name> 9000:9000

# 查看Pod详细信息
kubectl describe pod <pod-name>

# 查看资源使用
kubectl top pod <pod-name>
```

## 安全最佳实践

### 1. 使用Secrets管理敏感信息

```bash
# 不要在代码中硬编码密钥
# 使用Kubernetes Secrets或外部密钥管理服务（如Vault）
```

### 2. 限制容器权限

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
```

### 3. 网络策略

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: tanzanite-api-netpol
spec:
  podSelector:
    matchLabels:
      app: tanzanite-api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: nginx-ingress
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
  - to:
    - podSelector:
        matchLabels:
          app: redis
```

### 4. RBAC配置

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tanzanite-api
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: tanzanite-api-role
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: tanzanite-api-binding
subjects:
- kind: ServiceAccount
  name: tanzanite-api
roleRef:
  kind: Role
  name: tanzanite-api-role
  apiGroup: rbac.authorization.k8s.io
```

## 备份和恢复

### 数据库备份

```bash
# 使用CronJob自动备份
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
spec:
  schedule: "0 2 * * *"  # 每天凌晨2点
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:14
            command:
            - /bin/sh
            - -c
            - pg_dump -h $DB_HOST -U $DB_USER $DB_NAME > /backup/backup-$(date +%Y%m%d).sql
```

## 成本优化

### 1. 使用节点亲和性

将Pod调度到更便宜的节点：

```yaml
nodeSelector:
  node-type: spot
```

### 2. 配置PodDisruptionBudget

确保滚动更新期间的可用性：

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: tanzanite-api-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: tanzanite-api
```

### 3. 使用Spot实例

对于非关键负载，考虑使用Spot/Preemptible实例。

## 性能优化

### 1. 调整资源限制

根据实际使用情况调整requests和limits。

### 2. 启用连接池

确保数据库和Redis连接池配置合理：

```yaml
database:
  max_idle_conns: 10
  max_open_conns: 100
redis:
  pool_size: 20
```

### 3. 使用CDN

静态资源使用CDN加速：

```yaml
storage:
  s3:
    base_url: "https://cdn.tanzanite.example.com"
```

## 生产清单

部署到生产环境前的检查清单：

- [ ] 所有Secrets已正确配置
- [ ] 数据库迁移已执行
- [ ] 健康检查端点正常工作
- [ ] 限流配置已测试
- [ ] HPA已配置并测试
- [ ] 日志聚合已设置
- [ ] 监控和告警已配置
- [ ] 备份策略已实施
- [ ] SSL证书已配置
- [ ] DNS记录已更新
- [ ] 负载测试已通过
- [ ] 灾难恢复计划已制定
- [ ] 文档已更新

## 相关文档

- [API文档](../API.md)
- [配置指南](../config/README.md)
- [开发指南](../DEVELOPMENT.md)
- [故障排查](../TROUBLESHOOTING.md)

## 支持

如有问题，请联系：
- 技术支持: support@tanzanite.example.com
- 开发团队: dev@tanzanite.example.com
