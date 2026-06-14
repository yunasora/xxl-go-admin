/**
 * Frontend-consumable backend contracts.
 *
 * The Go backend currently returns XXL-style envelopes:
 * { code: number, msg?: string, content?: unknown }
 *
 * JobInfo JSON fields observed from models/job_info.go:
 * Id, jobGroup, jobDesc, executorHandler, jobCron, executorRoutingStrategy,
 * executorParam, executorTimeout, executorFailRetryCount, triggerStatus,
 * triggerLastTime, triggerNextTime, createTime, updateTime.
 *
 * Unsupported by the current backend API surface:
 * - job_group CRUD/list
 * - real job_log query
 * - executor registry list/detail/delete
 * - auth/session APIs
 */
export const supportedApiPaths = [
  'GET /api/ui-config',
  'POST /api/registry',
  'POST /api/job',
  'GET /api/job',
  'GET /api/job/:id',
  'PUT /api/job/:id',
  'DELETE /api/job/:id',
  'PUT /api/job/:id/start',
  'PUT /api/job/:id/stop',
];

export const missingApiCapabilities = [
  'job_group 管理接口',
  '真实 job_log 查询接口',
  '执行器在线列表接口',
  '登录、鉴权与用户会话接口',
  '服务端分页、筛选与排序接口',
];
