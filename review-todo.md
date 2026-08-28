# Code Review Follow-up TODO

> 本清单基于 `feat/beeper-apps-and-models` 分支的代码审查，整理待处理的优化项与跟进事项。

---

## 🛠 待处理项 (Action Items)

- [x] **1. 提交工作区未暂存的改动**
  - **位置**: `core/app/controllers/api/v1/beepers_controller.rb` & `core/test/controllers/api/v1/beepers_controller_test.rb`
  - **说明**: 将 `find_beeper_app` 查找范围从仅 `BeeperApp.official` 扩展为 `account_id: [nil, Current.account.id]`，使账号可安装私有/自定义 Beeper App。已添加集成测试验证。

- [x] **2. 优化 Beeper App 查找失败的错误提示文案**
  - **位置**: `core/app/controllers/api/v1/beepers_controller.rb:17`
  - **说明**: 当前文案已统一调整为 `"Beeper app not found"`。

- [x] **3. 增强 `BeeperRun#sanitize_signal_result` 的安全截断**
  - **位置**: `core/app/models/beeper_run.rb:62-94`
  - **说明**: 当 `signal_result` 超过 `8.kilobytes` 时，截断 `message`、`title` 并将 `metrics` 限制最多保留 20 项并对长值做截断；若仍然超限则清空 `metrics`，确保最终存储安全。已添加针对性单元测试。

- [x] **4. 完善探针 Cron 表达式最小频率限制 (Min Interval Validation)**
  - **位置**: `core/app/models/beeper.rb` & `core/app/models/beeper_app.rb`
  - **说明**: 联动 Manifest 的 `schedule.min_interval_seconds`，在 Beeper 校验 Cron 时，通过计算连续两次执行间隔，禁止配置低于探针所允许最小频率的 Cron 表达式。已添加针对性单元测试。
