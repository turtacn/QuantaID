# 设备信任管理 (Device Trust Management)

## 概述

设备信任管理模块旨在通过分析设备属性、行为模式和历史数据，为每个访问系统的设备建立信任评分。该评分可用于增强认证决策（如适应性MFA）、风险控制和异常检测。

## 核心组件

### 1. 数据模型 (Data Model)

设备实体 (`Device`) 包含以下关键信息：
- **ID**: 唯一标识符
- **Fingerprint**: 基于硬件和软件特征生成的唯一哈希
- **TrustScore**: 0-100 的信任分数
- **Binding**: 关联的用户 ID
- **Activity**: 最后活跃时间、IP 和位置

### 2. 指纹生成 (Fingerprinting)

使用 `DeviceFingerprinter` 生成一致的设备指纹。考虑的因素包括：
- User Agent
- Screen Resolution
- Timezone
- Platform
- Hardware Concurrency
- Canvas/WebGL Hash

### 3. 信任评分 (Trust Scoring)

`TrustScorer` 根据以下规则计算信任分：
- **基础分 (Base Score)**: 新设备的起始分
- **设备年龄 (Age Bonus)**: 注册时间越长，分数越高
- **绑定状态 (Bound Bonus)**: 绑定到特定用户后加分
- **验证加分 (Verified Bonus)**: 长期活跃且无异常的设备获得额外加分

### 4. 异常检测 (Anomaly Detection)

`AnomalyDetector` 实时监控设备行为：
- **地理位置跳跃 (Geo Jump)**: 检测短时间内不可能的物理移动
- **指纹突变 (Fingerprint Change)**: 检测核心设备特征的变更
- **异常时间 (Unusual Time)**: 识别非习惯性访问时间

## 配置

在 `server.yaml` 中配置 `security.device_trust`：

```yaml
security:
  device_trust:
    base_score: 20
    age_bonus: 1
    max_age_bonus: 30
    bound_bonus: 20
    verified_bonus: 20
    max_speed_kmh: 900
```

## API 使用

- **RegisterOrUpdate**: 在登录或访问时调用，注册新设备或更新现有设备信息。
- **BindToUser**: 用户登录成功后，将设备与用户绑定。
- **GetTrustLevel**: 获取当前设备的信任级别 (Low, Medium, High, Verified)。
