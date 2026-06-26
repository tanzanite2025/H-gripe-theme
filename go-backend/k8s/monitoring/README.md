# Monitoring Stack Setup

完整的Prometheus + Grafana + Alertmanager监控栈配置。

## 架构概览

```
┌─────────────────────────────────────────────────────────┐
│                    Monitoring Stack                      │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────┐    ┌──────────────┐    ┌───────────┐ │
│  │  Prometheus  │───▶│ Alertmanager │───▶│  Webhook  │ │
│  │   (Metrics)  │    │   (Alerts)   │    │  (Slack)  │ │
│  └──────────────┘    └──────────────┘    └───────────┘ │
│         │                                                │
│         │ scrape                                         │
│         ▼                                                │
│  ┌──────────────────────────────────────┐               │
│  │  Tanzanite API (Port 9000/metrics)   │               │
│  │  PostgreSQL Exporter (Port 9187)     │               │
│  │  Redis Exporter (Port 9121)          │               │
│  └──────────────────────────────────────┘               │
│         │                                                │
│         │ query                                          │
│         ▼                                                │
│  ┌──────────────┐                                       │
│  │   Grafana    │                                       │
│  │ (Dashboard)  │                                       │
│  └──────────────┘                                       │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

## 快速开始

### 1. 创建monitoring命名空间

```bash
kubectl create namespace monitoring
```

### 2. 部署Prometheus

```bash
# 应用Prometheus配置
kubectl apply -f prometheus-config.yaml

# 验证部署
kubectl get pods -n monitoring -l app=prometheus
kubectl logs -n monitoring -l app=prometheus
```

### 3. 部署Alertmanager

```bash
# 应用Alertmanager配置
kubectl apply -f alertmanager-config.yaml

# 验证部署
kubectl get pods -n monitoring -l app=alertmanager
```

### 4. 部署Grafana

```bash
# 更新Grafana密码
kubectl create secret generic grafana-secrets \
  --from-literal=admin-user=admin \
  --from-literal=admin-password=YOUR_SECURE_PASSWORD \
  -n monitoring

# 应用Grafana配置
kubectl apply -f grafana-config.yaml

# 验证部署
kubectl get pods -n monitoring -l app=grafana
```

### 5. 部署Exporters

```bash
# PostgreSQL Exporter
kubectl apply -f exporters/postgres-exporter.yaml

# Redis Exporter
kubectl apply -f exporters/redis-exporter.yaml
```

## 访问监控系统

### Prometheus UI

```bash
# 端口转发
kubectl port-forward -n monitoring svc/prometheus 9090:9090

# 访问
http://localhost:9090
```

### Grafana Dashboard

```bash
# 端口转发
kubectl port-forward -n monitoring svc/grafana 3000:3000

# 访问
http://localhost:3000

# 默认凭证（请立即修改）
用户名: admin
密码: (在Secret中设置的密码)
```

### Alertmanager UI

```bash
# 端口转发
kubectl port-forward -n monitoring svc/alertmanager 9093:9093

# 访问
http://localhost:9093
```

## 生产环境访问

### 通过Ingress访问（推荐）

配置文件中已包含Ingress配置，需要：

1. 更新域名：
   - `prometheus.tanzanite.example.com`
   - `grafana.tanzanite.example.com`
   - `alertmanager.tanzanite.example.com`

2. 确保DNS记录指向Ingress Controller

3. 配置TLS证书（使用cert-manager）

## 监控指标

### API指标

Tanzanite API暴露的Prometheus指标：

| 指标名 | 类型 | 描述 |
|--------|------|------|
| `http_requests_total` | Counter | HTTP请求总数 |
| `http_request_duration_seconds` | Histogram | 请求响应时间 |
| `http_requests_in_flight` | Gauge | 当前处理中的请求数 |
| `db_connections_in_use` | Gauge | 当前使用的数据库连接数 |
| `db_connections_max` | Gauge | 最大数据库连接数 |
| `cache_hits_total` | Counter | 缓存命中次数 |
| `cache_misses_total` | Counter | 缓存未命中次数 |

### 系统指标

通过Kubernetes和Node Exporter收集：

- CPU使用率
- 内存使用率
- 磁盘I/O
- 网络流量
- Pod状态和重启次数

### 数据库指标

通过PostgreSQL Exporter收集：

- 活动连接数
- 查询性能
- 慢查询
- 表大小
- 索引使用情况

### 缓存指标

通过Redis Exporter收集：

- 内存使用
- 命中率
- 连接数
- 键空间统计

## 告警规则

### 严重告警 (Critical)

1. **APIDown**: API服务不可用
2. **DatabaseConnectionPoolExhausted**: 数据库连接池耗尽
3. **RedisConnectionFailed**: Redis连接失败

### 警告告警 (Warning)

1. **HighErrorRate**: 错误率超过5%
2. **HighResponseTime**: P95响应时间超过1秒
3. **HighCPUUsage**: CPU使用率超过80%
4. **HighMemoryUsage**: 内存使用率超过80%
5. **PodRestartingTooOften**: Pod频繁重启

## Grafana仪表板

### 内置仪表板

1. **API Metrics**: API性能概览
   - 请求率
   - 错误率
   - 响应时间
   - 活动连接数

2. **Infrastructure**: 基础设施监控
   - CPU/内存使用
   - 网络流量
   - 磁盘I/O

3. **Database**: 数据库监控
   - 连接数
   - 查询性能
   - 慢查询日志

4. **Redis**: 缓存监控
   - 命中率
   - 内存使用
   - 键空间统计

### 导入自定义仪表板

1. 访问Grafana UI
2. 点击 "+" -> "Import"
3. 上传JSON文件或输入Dashboard ID
4. 选择Prometheus数据源

推荐仪表板：
- Node Exporter Full (ID: 1860)
- PostgreSQL Database (ID: 9628)
- Redis Dashboard (ID: 11835)
- Kubernetes Cluster Monitoring (ID: 7249)

## 告警通知配置

### Slack集成

编辑 `alertmanager-config.yaml`：

```yaml
receivers:
- name: 'slack'
  slack_configs:
  - api_url: 'YOUR_SLACK_WEBHOOK_URL'
    channel: '#alerts'
    title: 'Tanzanite Alert'
```

### 邮件通知

```yaml
receivers:
- name: 'email'
  email_configs:
  - to: 'ops@tanzanite.com'
    from: 'alertmanager@tanzanite.com'
    smarthost: 'smtp.gmail.com:587'
    auth_username: 'your-email@gmail.com'
    auth_password: 'your-app-password'
```

### PagerDuty集成

```yaml
receivers:
- name: 'pagerduty'
  pagerduty_configs:
  - service_key: 'YOUR_PAGERDUTY_SERVICE_KEY'
```

## 数据持久化

### Prometheus数据存储

默认使用emptyDir，生产环境建议使用PersistentVolume：

```yaml
volumes:
- name: prometheus-storage
  persistentVolumeClaim:
    claimName: prometheus-pvc
```

创建PVC：

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: prometheus-pvc
  namespace: monitoring
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
  storageClassName: standard
```

### Grafana数据存储

同样建议使用PersistentVolume保存仪表板配置。

## 性能调优

### Prometheus

1. **调整抓取间隔**：
   ```yaml
   global:
     scrape_interval: 30s  # 降低频率以节省资源
   ```

2. **增加保留时间**：
   ```yaml
   --storage.tsdb.retention.time=90d
   ```

3. **启用压缩**：
   ```yaml
   --storage.tsdb.wal-compression
   ```

### Grafana

1. **启用缓存**：
   ```ini
   [caching]
   enabled = true
   ```

2. **限制查询范围**：
   - 使用合适的时间范围
   - 避免长时间范围的高精度查询

## 安全建议

1. **更改默认密码**：
   - Grafana admin密码
   - Alertmanager web UI密码

2. **启用HTTPS**：
   - 使用cert-manager自动管理证书
   - 强制SSL重定向

3. **网络策略**：
   ```yaml
   apiVersion: networking.k8s.io/v1
   kind: NetworkPolicy
   metadata:
     name: monitoring-network-policy
     namespace: monitoring
   spec:
     podSelector:
       matchLabels:
         app: prometheus
     policyTypes:
     - Ingress
     ingress:
     - from:
       - namespaceSelector:
           matchLabels:
             name: production
   ```

4. **RBAC权限控制**：
   - 限制ServiceAccount权限
   - 使用最小权限原则

## 故障排查

### Prometheus无法抓取指标

```bash
# 检查目标状态
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# 访问 http://localhost:9090/targets

# 检查网络连接
kubectl exec -n monitoring deployment/prometheus -- wget -O- http://tanzanite-api:9000/metrics

# 查看日志
kubectl logs -n monitoring -l app=prometheus --tail=100
```

### Grafana无法连接Prometheus

```bash
# 测试连接
kubectl exec -n monitoring deployment/grafana -- wget -O- http://prometheus:9090/api/v1/query?query=up

# 检查数据源配置
kubectl describe configmap grafana-datasources -n monitoring
```

### 告警未触发

```bash
# 检查告警规则
kubectl exec -n monitoring deployment/prometheus -- promtool check rules /etc/prometheus/rules/alerts.yml

# 查看活动告警
# 访问 http://localhost:9090/alerts
```

## 备份和恢复

### 备份Prometheus数据

```bash
# 创建快照
curl -XPOST http://localhost:9090/api/v1/admin/tsdb/snapshot

# 复制数据
kubectl cp monitoring/prometheus-pod:/prometheus/snapshots/snapshot-name ./backup/
```

### 备份Grafana仪表板

```bash
# 导出所有仪表板
kubectl exec -n monitoring deployment/grafana -- grafana-cli admin export-dashboard
```

## 监控最佳实践

1. **设置合理的告警阈值**：避免告警疲劳
2. **使用标签分类**：便于过滤和聚合
3. **记录SLI/SLO**：服务水平指标和目标
4. **定期审查仪表板**：删除过时的面板
5. **监控监控系统**：确保Prometheus自身健康

## 扩展阅读

- [Prometheus官方文档](https://prometheus.io/docs/)
- [Grafana文档](https://grafana.com/docs/)
- [Alertmanager配置](https://prometheus.io/docs/alerting/latest/configuration/)
- [PromQL查询语言](https://prometheus.io/docs/prometheus/latest/querying/basics/)

## 支持

如有问题，请联系DevOps团队或查看：
- GitHub Issues
- Slack #devops频道
