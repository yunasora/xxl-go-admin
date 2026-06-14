import { useEffect, useMemo, useState } from 'react';
import {
  createJob,
  deleteJob,
  getJob,
  getUiConfig,
  listJobs,
  registerExecutor,
  startJob,
  stopJob,
  updateJob,
} from './api.js';
import { missingApiCapabilities, supportedApiPaths } from './types.js';

const navItems = [
  { key: 'dashboard', label: '概览仪表盘', badge: 'Home' },
  { key: 'jobs', label: '任务管理', badge: 'CRUD' },
  { key: 'executors', label: '执行器注册', badge: 'HTTP' },
  { key: 'logs', label: '日志/历史', badge: 'No API' },
  { key: 'settings', label: '接口边界', badge: 'Guide' },
];

const defaultJobForm = {
  jobGroup: '1',
  jobDesc: '',
  executorHandler: '',
  jobCron: '*/10 * * * * *',
  executorParam: '',
  executorTimeout: '10',
  executorFailRetryCount: '1',
  triggerStatus: '1',
};

const defaultRegistryForm = {
  registryGroup: 'EXECUTOR',
  registryKey: 'xxl-job-executor-sample',
  registryValue: 'http://127.0.0.1:9999/',
};

function formatDate(value) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '-';
  return date.toLocaleString('zh-CN');
}

function formatMillis(value) {
  if (!value) return '-';
  const date = new Date(Number(value));
  if (Number.isNaN(date.getTime())) return String(value);
  return date.toLocaleString('zh-CN');
}

function toPayload(form) {
  return {
    jobGroup: Number(form.jobGroup),
    jobDesc: form.jobDesc.trim(),
    executorHandler: form.executorHandler.trim(),
    jobCron: form.jobCron.trim(),
    executorParam: form.executorParam,
    executorTimeout: Number(form.executorTimeout || 0),
    executorFailRetryCount: Number(form.executorFailRetryCount || 0),
    triggerStatus: Number(form.triggerStatus || 0),
    triggerLastTime: 0,
    triggerNextTime: 0,
  };
}

function validateJobForm(form) {
  const payload = toPayload(form);
  if (!Number.isFinite(payload.jobGroup) || payload.jobGroup <= 0) return '任务分组必须是大于 0 的数字';
  if (!payload.jobDesc) return '任务描述不能为空';
  if (!payload.executorHandler) return 'Executor Handler 不能为空';
  if (!payload.jobCron) return 'Cron 表达式不能为空';
  if (!Number.isFinite(payload.executorTimeout) || payload.executorTimeout < 0) return '超时时间不能小于 0';
  if (!Number.isFinite(payload.executorFailRetryCount) || payload.executorFailRetryCount < 0) return '失败重试次数不能小于 0';
  return '';
}

function StatusPill({ running }) {
  return <span className={`status-pill ${running ? 'running' : 'stopped'}`}>{running ? '运行中' : '已停用'}</span>;
}

function EmptyState({ title, description, children }) {
  return (
    <div className="empty-state">
      <div className="empty-icon">∅</div>
      <h3>{title}</h3>
      <p>{description}</p>
      {children}
    </div>
  );
}

function App() {
  const [activeView, setActiveView] = useState('dashboard');
  const [theme, setTheme] = useState(() => localStorage.getItem('go-xxl-admin-theme') || 'dark');
  const [jobs, setJobs] = useState([]);
  const [uiConfig, setUiConfig] = useState(null);
  const [loading, setLoading] = useState(true);
  const [actionBusy, setActionBusy] = useState('');
  const [error, setError] = useState('');
  const [toast, setToast] = useState('');
  const [lastRefresh, setLastRefresh] = useState(null);
  const [filters, setFilters] = useState({ search: '', status: 'all', sort: 'id_desc', pageSize: 10, page: 1 });
  const [jobForm, setJobForm] = useState(defaultJobForm);
  const [editJob, setEditJob] = useState(null);
  const [detailJob, setDetailJob] = useState(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [registryForm, setRegistryForm] = useState(defaultRegistryForm);

  useEffect(() => {
    document.body.dataset.theme = theme;
    localStorage.setItem('go-xxl-admin-theme', theme);
  }, [theme]);

  useEffect(() => {
    refreshAll();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!toast) return undefined;
    const timer = window.setTimeout(() => setToast(''), 2600);
    return () => window.clearTimeout(timer);
  }, [toast]);

  const stats = useMemo(() => {
    const running = jobs.filter((job) => job.triggerStatus === 1).length;
    return {
      total: jobs.length,
      running,
      stopped: jobs.length - running,
    };
  }, [jobs]);

  const filteredJobs = useMemo(() => {
    const keyword = filters.search.trim().toLowerCase();
    let rows = [...jobs];
    if (keyword) {
      rows = rows.filter((job) => [job.id, job.jobGroup, job.jobDesc, job.executorHandler, job.jobCron]
        .some((value) => String(value ?? '').toLowerCase().includes(keyword)));
    }
    if (filters.status === 'running') rows = rows.filter((job) => job.triggerStatus === 1);
    if (filters.status === 'stopped') rows = rows.filter((job) => job.triggerStatus !== 1);
    rows.sort((a, b) => {
      if (filters.sort === 'id_asc') return a.id - b.id;
      if (filters.sort === 'update_desc') return new Date(b.updateTime || 0) - new Date(a.updateTime || 0);
      if (filters.sort === 'create_desc') return new Date(b.createTime || 0) - new Date(a.createTime || 0);
      return b.id - a.id;
    });
    return rows;
  }, [jobs, filters.search, filters.sort, filters.status]);

  const totalPages = Math.max(1, Math.ceil(filteredJobs.length / filters.pageSize));
  const currentPage = Math.min(filters.page, totalPages);
  const pagedJobs = filteredJobs.slice((currentPage - 1) * filters.pageSize, currentPage * filters.pageSize);

  async function refreshAll() {
    setLoading(true);
    setError('');
    try {
      const [config, rows] = await Promise.all([getUiConfig(), listJobs()]);
      setUiConfig(config);
      setJobs(rows);
      setLastRefresh(new Date());
    } catch (err) {
      setError(err.message || '加载失败');
    } finally {
      setLoading(false);
    }
  }

  async function refreshJobsOnly() {
    const rows = await listJobs();
    setJobs(rows);
    setLastRefresh(new Date());
  }

  function updateFilters(patch) {
    setFilters((prev) => ({ ...prev, ...patch, page: patch.page ?? 1 }));
  }

  function updateJobForm(patch) {
    setJobForm((prev) => ({ ...prev, ...patch }));
  }

  function updateEditForm(patch) {
    setEditJob((prev) => ({ ...prev, form: { ...prev.form, ...patch } }));
  }

  async function runAction(label, fn) {
    setActionBusy(label);
    setError('');
    try {
      await fn();
    } catch (err) {
      setError(err.message || '操作失败');
    } finally {
      setActionBusy('');
    }
  }

  async function handleCreate(event) {
    event.preventDefault();
    const validation = validateJobForm(jobForm);
    if (validation) {
      setError(validation);
      return;
    }
    await runAction('create', async () => {
      await createJob(toPayload(jobForm));
      setToast('任务创建成功，已重新拉取后端状态');
      setJobForm(defaultJobForm);
      await refreshJobsOnly();
    });
  }

  async function handleEditSubmit(event) {
    event.preventDefault();
    if (!editJob) return;
    const validation = validateJobForm(editJob.form);
    if (validation) {
      setError(validation);
      return;
    }
    await runAction(`edit-${editJob.id}`, async () => {
      await updateJob(editJob.id, toPayload(editJob.form));
      setEditJob(null);
      setToast('任务已更新，页面展示为后端重新拉取后的结果');
      await refreshJobsOnly();
    });
  }

  async function openDetail(job) {
    setDetailJob(job);
    setDetailLoading(true);
    try {
      const latest = await getJob(job.id);
      setDetailJob(latest);
    } catch (err) {
      setError(`详情刷新失败，已显示列表缓存：${err.message}`);
    } finally {
      setDetailLoading(false);
    }
  }

  function openEdit(job) {
    setEditJob({
      id: job.id,
      form: {
        jobGroup: String(job.jobGroup || ''),
        jobDesc: job.jobDesc || '',
        executorHandler: job.executorHandler || '',
        jobCron: job.jobCron || '',
        executorParam: job.executorParam || '',
        executorTimeout: String(job.executorTimeout ?? 0),
        executorFailRetryCount: String(job.executorFailRetryCount ?? 0),
        triggerStatus: String(job.triggerStatus ?? 0),
      },
    });
  }

  async function toggleJob(job) {
    await runAction(`toggle-${job.id}`, async () => {
      if (job.triggerStatus === 1) {
        await stopJob(job.id);
        setToast(`任务 #${job.id} 已停用`);
      } else {
        await startJob(job.id);
        setToast(`任务 #${job.id} 已启用`);
      }
      await refreshJobsOnly();
    });
  }

  async function removeJob(job) {
    if (!window.confirm(`确认删除任务 #${job.id} 吗？此操作不可在前端恢复。`)) return;
    await runAction(`delete-${job.id}`, async () => {
      await deleteJob(job.id);
      if (detailJob?.id === job.id) setDetailJob(null);
      setToast(`任务 #${job.id} 已删除`);
      await refreshJobsOnly();
    });
  }

  async function handleRegistrySubmit(event) {
    event.preventDefault();
    if (!registryForm.registryKey.trim() || !registryForm.registryValue.trim()) {
      setError('registryKey 和 registryValue 不能为空');
      return;
    }
    await runAction('registry', async () => {
      await registerExecutor({
        registryGroup: registryForm.registryGroup.trim() || 'EXECUTOR',
        registryKey: registryForm.registryKey.trim(),
        registryValue: registryForm.registryValue.trim(),
      });
      setToast('注册/心跳请求已提交成功');
    });
  }

  function renderContent() {
    if (activeView === 'dashboard') {
      return <Dashboard stats={stats} jobs={jobs} uiConfig={uiConfig} lastRefresh={lastRefresh} loading={loading} setActiveView={setActiveView} openDetail={openDetail} />;
    }
    if (activeView === 'jobs') {
      return (
        <JobsView
          jobs={pagedJobs}
          allCount={filteredJobs.length}
          filters={{ ...filters, page: currentPage }}
          totalPages={totalPages}
          form={jobForm}
          loading={loading}
          actionBusy={actionBusy}
          onFilterChange={updateFilters}
          onFormChange={updateJobForm}
          onCreate={handleCreate}
          onRefresh={refreshAll}
          onOpenDetail={openDetail}
          onOpenEdit={openEdit}
          onToggle={toggleJob}
          onDelete={removeJob}
        />
      );
    }
    if (activeView === 'executors') {
      return <ExecutorsView form={registryForm} setForm={setRegistryForm} onSubmit={handleRegistrySubmit} busy={actionBusy === 'registry'} />;
    }
    if (activeView === 'logs') {
      return <UnavailableLogsView />;
    }
    return <SettingsView uiConfig={uiConfig} />;
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark">GO</div>
          <div>
            <div className="brand-title">XXL Admin</div>
            <div className="brand-subtitle">React Console</div>
          </div>
        </div>
        <nav className="nav-list" aria-label="主导航">
          {navItems.map((item) => (
            <button key={item.key} type="button" className={`nav-item ${activeView === item.key ? 'active' : ''}`} onClick={() => setActiveView(item.key)}>
              <span>{item.label}</span>
              <span className="nav-badge">{item.key === 'dashboard' ? stats.total : item.badge}</span>
            </button>
          ))}
        </nav>
      </aside>

      <main className="main-content">
        <header className="topbar">
          <div>
            <p className="eyebrow">GO-XXL-ADMIN</p>
            <h1>调度中心前端控制台</h1>
            <p className="subtitle">React + Vite 实现，只接入当前后端已存在的真实接口。</p>
          </div>
          <div className="topbar-actions">
            <button type="button" className="btn secondary" onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>{theme === 'dark' ? '浅色主题' : '深色主题'}</button>
            <button type="button" className="btn primary" onClick={refreshAll} disabled={loading || !!actionBusy}>{loading ? '刷新中...' : '刷新全部'}</button>
          </div>
        </header>

        {error && <div className="alert error"><strong>操作提示：</strong>{error}<button type="button" onClick={() => setError('')}>关闭</button></div>}
        {renderContent()}
      </main>

      {detailJob && <DetailDrawer job={detailJob} loading={detailLoading} onClose={() => setDetailJob(null)} />}
      {editJob && <EditModal editJob={editJob} busy={actionBusy === `edit-${editJob.id}`} onChange={updateEditForm} onSubmit={handleEditSubmit} onClose={() => setEditJob(null)} />}
      {toast && <div className="toast">{toast}</div>}
    </div>
  );
}

function Dashboard({ stats, jobs, uiConfig, lastRefresh, loading, setActiveView, openDetail }) {
  const recent = [...jobs].sort((a, b) => b.id - a.id).slice(0, 6);
  return (
    <section className="view-stack">
      <div className="kpi-grid">
        <KpiCard label="任务总数" value={stats.total} foot="来源：GET /api/job" />
        <KpiCard label="运行中" value={stats.running} foot="triggerStatus = 1" tone="success" />
        <KpiCard label="已停用" value={stats.stopped} foot="triggerStatus != 1" tone="warning" />
        <KpiCard label="最后刷新" value={lastRefresh ? lastRefresh.toLocaleTimeString('zh-CN') : '--'} foot="浏览器本地时间" compact />
      </div>

      <div className="content-grid two-one">
        <section className="panel">
          <div className="panel-header">
            <div>
              <h2>最近任务</h2>
              <p>展示后端返回的最新任务数据，不做服务端能力假设。</p>
            </div>
            <button type="button" className="btn secondary" onClick={() => setActiveView('jobs')}>进入任务管理</button>
          </div>
          {loading ? <LoadingRows /> : recent.length ? <MiniJobTable jobs={recent} openDetail={openDetail} /> : <EmptyState title="暂无任务" description="后端当前没有返回任务数据，请先创建任务或准备数据库数据。" />}
        </section>

        <section className="panel">
          <div className="panel-header compact"><h2>后端运行配置</h2></div>
          <div className="info-list">
            <InfoRow label="服务端口" value={uiConfig?.serverPort || '-'} />
            <InfoRow label="AppName" value={uiConfig?.appName || '未配置'} />
            <InfoRow label="Gin 模式" value={uiConfig?.ginMode || '-'} />
            <InfoRow label="HTTP 超时" value={`${uiConfig?.httpTimeout ?? '-'}s`} />
            <InfoRow label="调度扫描" value={`${uiConfig?.schedulerInterval ?? '-'}s`} />
            <InfoRow label="注册超时" value={`${uiConfig?.registryTimeout ?? '-'}s`} />
            <InfoRow label="RabbitMQ" value={uiConfig?.mqEnabled ? '开启' : '关闭'} />
            <InfoRow label="Redis" value={uiConfig?.redisEnabled ? '开启' : '关闭'} />
          </div>
          <div className="boundary-note">执行器在线数、真实日志与 job_group 列表当前没有后端查询接口，因此这里不会伪造展示。</div>
        </section>
      </div>
    </section>
  );
}

function KpiCard({ label, value, foot, tone = '', compact = false }) {
  return (
    <div className={`kpi-card ${tone}`}>
      <span>{label}</span>
      <strong className={compact ? 'compact' : ''}>{value}</strong>
      <small>{foot}</small>
    </div>
  );
}

function JobsView({ jobs, allCount, filters, totalPages, form, loading, actionBusy, onFilterChange, onFormChange, onCreate, onRefresh, onOpenDetail, onOpenEdit, onToggle, onDelete }) {
  const start = allCount ? (filters.page - 1) * filters.pageSize + 1 : 0;
  const end = Math.min(allCount, filters.page * filters.pageSize);
  return (
    <section className="content-grid job-layout">
      <form className="panel form-panel" onSubmit={onCreate}>
        <div className="panel-header compact">
          <div>
            <h2>创建任务</h2>
            <p>创建时仅提交当前后端 JobInfo 支持的字段。</p>
          </div>
        </div>
        <FormNote text="job_group 目前没有前端可用管理接口，请先由后端/数据库侧准备有效分组 ID。" />
        <FormNote tone="warn" text="当前后端会把创建时 triggerStatus=0 的任务改为启用状态，前端默认按启用提交。" />
        <JobFormFields form={form} onChange={onFormChange} />
        <button type="submit" className="btn primary wide" disabled={actionBusy === 'create'}>{actionBusy === 'create' ? '创建中...' : '创建任务'}</button>
      </form>

      <section className="panel table-panel">
        <div className="panel-header">
          <div>
            <h2>任务列表</h2>
            <p>客户端筛选、排序和分页；后端当前返回全量任务。</p>
          </div>
          <button type="button" className="btn secondary" onClick={onRefresh} disabled={loading || !!actionBusy}>刷新</button>
        </div>
        <div className="filters">
          <input value={filters.search} onChange={(e) => onFilterChange({ search: e.target.value })} placeholder="搜索 ID / 描述 / Handler / Cron / 分组" />
          <select value={filters.status} onChange={(e) => onFilterChange({ status: e.target.value })}>
            <option value="all">全部状态</option>
            <option value="running">运行中</option>
            <option value="stopped">已停用</option>
          </select>
          <select value={filters.sort} onChange={(e) => onFilterChange({ sort: e.target.value })}>
            <option value="id_desc">ID 从新到旧</option>
            <option value="id_asc">ID 从旧到新</option>
            <option value="update_desc">更新时间倒序</option>
            <option value="create_desc">创建时间倒序</option>
          </select>
          <select value={filters.pageSize} onChange={(e) => onFilterChange({ pageSize: Number(e.target.value) })}>
            <option value={10}>10 条/页</option>
            <option value={20}>20 条/页</option>
            <option value={50}>50 条/页</option>
          </select>
        </div>

        {loading ? <LoadingRows /> : jobs.length ? (
          <>
            <div className="table-wrap">
              <table>
                <thead><tr><th>ID</th><th>分组</th><th>描述</th><th>Handler</th><th>Cron</th><th>超时</th><th>重试</th><th>更新时间</th><th>状态</th><th>操作</th></tr></thead>
                <tbody>{jobs.map((job) => (
                  <tr key={job.id}>
                    <td>#{job.id}</td>
                    <td>{job.jobGroup || '-'}</td>
                    <td className="strong-cell">{job.jobDesc || '-'}</td>
                    <td><code>{job.executorHandler || '-'}</code></td>
                    <td><code>{job.jobCron || '-'}</code></td>
                    <td>{job.executorTimeout}s</td>
                    <td>{job.executorFailRetryCount}</td>
                    <td>{formatDate(job.updateTime)}</td>
                    <td><StatusPill running={job.triggerStatus === 1} /></td>
                    <td>
                      <div className="row-actions">
                        <button type="button" className="btn small secondary" onClick={() => onOpenDetail(job)}>详情</button>
                        <button type="button" className="btn small secondary" onClick={() => onOpenEdit(job)}>编辑</button>
                        <button type="button" className={`btn small ${job.triggerStatus === 1 ? 'warning' : 'success'}`} disabled={actionBusy === `toggle-${job.id}`} onClick={() => onToggle(job)}>{job.triggerStatus === 1 ? '停用' : '启用'}</button>
                        <button type="button" className="btn small danger" disabled={actionBusy === `delete-${job.id}`} onClick={() => onDelete(job)}>删除</button>
                      </div>
                    </td>
                  </tr>
                ))}</tbody>
              </table>
            </div>
            <div className="pagination">
              <span>第 {filters.page} / {totalPages} 页 · 显示 {start}-{end} / {allCount}</span>
              <div className="page-actions">
                <button type="button" className="btn small secondary" disabled={filters.page <= 1} onClick={() => onFilterChange({ page: filters.page - 1 })}>上一页</button>
                <button type="button" className="btn small secondary" disabled={filters.page >= totalPages} onClick={() => onFilterChange({ page: filters.page + 1 })}>下一页</button>
              </div>
            </div>
          </>
        ) : <EmptyState title="没有匹配任务" description="请调整筛选条件，或创建一条新任务。" />}
      </section>
    </section>
  );
}

function JobFormFields({ form, onChange }) {
  return (
    <div className="form-grid">
      <label>任务分组 ID<input value={form.jobGroup} onChange={(e) => onChange({ jobGroup: e.target.value })} inputMode="numeric" /></label>
      <label>任务描述<input value={form.jobDesc} onChange={(e) => onChange({ jobDesc: e.target.value })} placeholder="例如：订单超时扫描" /></label>
      <label>Executor Handler<input value={form.executorHandler} onChange={(e) => onChange({ executorHandler: e.target.value })} placeholder="demoJobHandler" /></label>
      <label>Cron<input value={form.jobCron} onChange={(e) => onChange({ jobCron: e.target.value })} placeholder="*/10 * * * * *" /></label>
      <label>超时秒数<input value={form.executorTimeout} onChange={(e) => onChange({ executorTimeout: e.target.value })} inputMode="numeric" /></label>
      <label>失败重试<input value={form.executorFailRetryCount} onChange={(e) => onChange({ executorFailRetryCount: e.target.value })} inputMode="numeric" /></label>
      <label>初始状态<select value={form.triggerStatus} onChange={(e) => onChange({ triggerStatus: e.target.value })}><option value="1">启用</option><option value="0">停用（当前后端创建时会改为启用）</option></select></label>
      <label className="full">执行参数<textarea value={form.executorParam} onChange={(e) => onChange({ executorParam: e.target.value })} placeholder="传给 Executor 的参数" /></label>
    </div>
  );
}

function ExecutorsView({ form, setForm, onSubmit, busy }) {
  return (
    <section className="content-grid two-one">
      <form className="panel" onSubmit={onSubmit}>
        <div className="panel-header compact"><div><h2>执行器注册</h2><p>调用真实后端接口 POST /api/registry。</p></div></div>
        <div className="form-grid single">
          <label>Registry Group<input value={form.registryGroup} onChange={(e) => setForm((prev) => ({ ...prev, registryGroup: e.target.value }))} /></label>
          <label>Registry Key / AppName<input value={form.registryKey} onChange={(e) => setForm((prev) => ({ ...prev, registryKey: e.target.value }))} /></label>
          <label>Registry Value / Executor 地址<input value={form.registryValue} onChange={(e) => setForm((prev) => ({ ...prev, registryValue: e.target.value }))} /></label>
        </div>
        <FormNote text="当前后端只提供注册/心跳写入接口，没有执行器列表查询接口；提交成功不代表前端可以读取在线节点列表。" />
        <button type="submit" className="btn primary wide" disabled={busy}>{busy ? '提交中...' : '提交注册/心跳'}</button>
      </form>
      <section className="panel">
        <div className="panel-header compact"><h2>当前边界</h2></div>
        <EmptyState title="暂无执行器列表" description="后端尚未提供 registry/executor 查询接口，所以前端不会伪造在线节点数据。" />
      </section>
    </section>
  );
}

function UnavailableLogsView() {
  return (
    <section className="panel centered-panel">
      <EmptyState title="当前后端暂无 job_log 查询接口" description="前端不会展示模拟执行日志，以避免把本地数据误认为真实调度结果。">
        <div className="api-hint">
          <strong>后续如果你自己补后端，可以考虑提供：</strong>
          <code>GET /api/job-log</code>
          <code>GET /api/job/:id/logs</code>
          <code>GET /api/job-log/:id</code>
        </div>
      </EmptyState>
    </section>
  );
}

function SettingsView({ uiConfig }) {
  return (
    <section className="content-grid two-one">
      <div className="panel">
        <div className="panel-header compact"><h2>已接入的真实接口</h2></div>
        <div className="api-list">{supportedApiPaths.map((item) => <code key={item}>{item}</code>)}</div>
      </div>
      <div className="panel">
        <div className="panel-header compact"><h2>暂不假装存在的能力</h2></div>
        <ul className="boundary-list">{missingApiCapabilities.map((item) => <li key={item}>{item}</li>)}</ul>
        <div className="boundary-note">这些能力等你后端实现后，前端再按真实接口接入。</div>
      </div>
      <div className="panel full-span">
        <div className="panel-header compact"><h2>当前 UI Config</h2></div>
        <pre className="json-block">{JSON.stringify(uiConfig || {}, null, 2)}</pre>
      </div>
    </section>
  );
}

function DetailDrawer({ job, loading, onClose }) {
  return (
    <div className="overlay" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <aside className="drawer" role="dialog" aria-modal="true" aria-label="任务详情">
        <div className="drawer-header"><div><p className="eyebrow">Job Detail</p><h2>任务 #{job.id}</h2></div><button type="button" className="icon-btn" onClick={onClose}>×</button></div>
        {loading && <div className="inline-loading">正在拉取最新详情...</div>}
        <div className="info-list large">
          <InfoRow label="任务描述" value={job.jobDesc || '-'} />
          <InfoRow label="任务分组" value={job.jobGroup || '-'} />
          <InfoRow label="Handler" value={job.executorHandler || '-'} mono />
          <InfoRow label="Cron" value={job.jobCron || '-'} mono />
          <InfoRow label="状态" value={<StatusPill running={job.triggerStatus === 1} />} />
          <InfoRow label="超时" value={`${job.executorTimeout}s`} />
          <InfoRow label="失败重试" value={job.executorFailRetryCount} />
          <InfoRow label="上次触发" value={formatMillis(job.triggerLastTime)} />
          <InfoRow label="下次触发" value={formatMillis(job.triggerNextTime)} />
          <InfoRow label="创建时间" value={formatDate(job.createTime)} />
          <InfoRow label="更新时间" value={formatDate(job.updateTime)} />
        </div>
        <h3>执行参数</h3>
        <pre className="json-block">{job.executorParam || '(空)'}</pre>
      </aside>
    </div>
  );
}

function EditModal({ editJob, busy, onChange, onSubmit, onClose }) {
  return (
    <div className="overlay centered" role="presentation" onMouseDown={(event) => event.target === event.currentTarget && onClose()}>
      <form className="modal" role="dialog" aria-modal="true" aria-label="编辑任务" onSubmit={onSubmit}>
        <div className="drawer-header"><div><p className="eyebrow">Edit Job</p><h2>编辑任务 #{editJob.id}</h2></div><button type="button" className="icon-btn" onClick={onClose}>×</button></div>
        <FormNote tone="warn" text="保存后会重新读取后端状态；如果 0/空值未生效，说明当前后端更新逻辑尚未支持该零值变更。" />
        <JobFormFields form={editJob.form} onChange={onChange} />
        <div className="modal-actions"><button type="button" className="btn secondary" onClick={onClose}>取消</button><button type="submit" className="btn primary" disabled={busy}>{busy ? '保存中...' : '保存修改'}</button></div>
      </form>
    </div>
  );
}

function MiniJobTable({ jobs, openDetail }) {
  return (
    <div className="table-wrap compact-table">
      <table>
        <thead><tr><th>ID</th><th>描述</th><th>Handler</th><th>Cron</th><th>状态</th><th></th></tr></thead>
        <tbody>{jobs.map((job) => <tr key={job.id}><td>#{job.id}</td><td>{job.jobDesc || '-'}</td><td><code>{job.executorHandler || '-'}</code></td><td><code>{job.jobCron || '-'}</code></td><td><StatusPill running={job.triggerStatus === 1} /></td><td><button type="button" className="btn small secondary" onClick={() => openDetail(job)}>详情</button></td></tr>)}</tbody>
      </table>
    </div>
  );
}

function InfoRow({ label, value, mono = false }) {
  return <div className="info-row"><span>{label}</span><strong className={mono ? 'mono' : ''}>{value}</strong></div>;
}

function FormNote({ text, tone = 'info' }) {
  return <div className={`form-note ${tone}`}>{text}</div>;
}

function LoadingRows() {
  return <div className="loading-card"><span className="spinner" /> 正在读取后端数据...</div>;
}

export default App;
