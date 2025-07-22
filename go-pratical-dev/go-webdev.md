# 1、net/http包

有能力的推荐去看官方文档，不过都是英文：https://golang.google.cn/ref/spec
简单学习的话可以看一下golang爱好者用中文写的文档：https://golang.halfiisland.com/essential/std/http.html#%E5%A2%9E%E5%8A%A0-header

# 2、开发框架-gin

推荐阅读官方文档：https://gin-gonic.com/zh-cn/docs/

# 3、数据库操作-gorm

推荐阅读官方文档：https://gorm.io/zh_CN/docs/

# 4. 认证授权
   1、jwt认证方式（有很多优秀的jwt库可供选择，这里选择了一个常用的）：

   推荐ai+官方文档一起使用：https://jwt.io/introduction

   2、OAuth2.0认证方式（同理，选了个用的人数很多的一个库）：
   推荐阅读官方文档：
   OAuth2.0：golang.org/x/oauth2

# 5. 数据验证
   学习文档链接：
库推荐：
     github.com/go-playground/validator/v10
 

# 6. 日志处理

库推荐：
   go.uber.org/zap：高性能结构化日志
   github.com/sirupsen/logrus：功能丰富的日志库

# 7. 缓存系统

   Redis 客户端：github.com/go-redis/redis/v8
   缓存热点数据，减轻数据库压力
   实现分布式锁、计数器、排行榜
   支持连接池配置和集群模式
   本地缓存：github.com/patrickmn/go-cache
   应用内内存缓存，适合单机场景
   支持过期时间和淘汰策略

# 8. 消息队列

   RabbitMQ：github.com/streadway/amqp
   可靠的消息投递，支持多种交换机类型
   适合业务解耦、异步处理（如邮件发送）
   Kafka：github.com/segmentio/kafka-go
   高吞吐量，适合日志收集、大数据场景
   分布式流处理平台

# 9. 配置管理

   github.com/spf13/viper
   支持多种格式（JSON, YAML, TOML 等）
   环境变量覆盖和配置热更新
   配置项默认值和校验

# 10. 文档生成

    github.com/swaggo/gin-swagger
    基于注释自动生成 Swagger API 文档
    支持接口测试和参数说明
    方便前后端协作

# 11. 任务调度

    github.com/robfig/cron/v3
    定时任务管理（如每天凌晨备份数据）
    支持秒级精度和复杂表达式

# 12. 监控告警

    Prometheus 客户端：github.com/prometheus/client_golang/prometheus
    暴露指标接口，监控系统运行状态
    配合 Grafana 可视化监控数据
    链路追踪：github.com/openzipkin/zipkin-go
    分布式系统调用链追踪
    定位性能瓶颈和服务依赖


# 13. 测试工具

    内置testing包：单元测试和基准测试
    github.com/stretchr/testify：断言库，简化测试代码
    github.com/gavv/httpexpect：HTTP 接口测试
