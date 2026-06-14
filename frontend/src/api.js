const JSON_HEADERS = { 'Content-Type': 'application/json' };

async function request(path, options = {}) {
  let response;
  try {
    response = await fetch(path, {
      ...options,
      headers: {
        ...(options.body ? JSON_HEADERS : {}),
        ...(options.headers || {}),
      },
    });
  } catch (error) {
    throw new Error(`网络请求失败：${error.message}`);
  }

  let data;
  try {
    data = await response.json();
  } catch {
    throw new Error(`接口返回非 JSON：${path}`);
  }

  if (data.code !== 200) {
    throw new Error(data.msg || `请求失败：${path}`);
  }

  return data.content;
}

export function normalizeJob(job = {}) {
  const id = Number(job.Id ?? job.id ?? 0);
  return {
    ...job,
    id,
    Id: job.Id ?? id,
    jobGroup: Number(job.jobGroup ?? 0),
    executorTimeout: Number(job.executorTimeout ?? 0),
    executorFailRetryCount: Number(job.executorFailRetryCount ?? 0),
    triggerStatus: Number(job.triggerStatus ?? 0),
    triggerLastTime: Number(job.triggerLastTime ?? 0),
    triggerNextTime: Number(job.triggerNextTime ?? 0),
  };
}

export function getUiConfig() {
  return request('/api/ui-config');
}

export async function listJobs() {
  const jobs = await request('/api/job');
  return Array.isArray(jobs) ? jobs.map(normalizeJob) : [];
}

export async function getJob(id) {
  const job = await request(`/api/job/${id}`);
  return normalizeJob(job);
}

export function createJob(payload) {
  return request('/api/job', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateJob(id, payload) {
  return request(`/api/job/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function deleteJob(id) {
  return request(`/api/job/${id}`, { method: 'DELETE' });
}

export function startJob(id) {
  return request(`/api/job/${id}/start`, { method: 'PUT' });
}

export function stopJob(id) {
  return request(`/api/job/${id}/stop`, { method: 'PUT' });
}

export function registerExecutor(payload) {
  return request('/api/registry', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}
