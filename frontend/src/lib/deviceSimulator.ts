export type DeviceConfig = {
  deviceId: string;
  group: string;
  productModel: string;
  hardwareVersion: string;
  currentVersion: string;
  apiToken: string;
};

type ApiEnvelope<T = unknown> = {
  code: number;
  message: string;
  data: T;
};

export type CheckUpdateData = {
  has_update?: boolean;
  task_id?: string;
  target_version?: string;
  download_url?: string;
  reason?: string;
  retry_after_sec?: number;
};

async function devicePost<T>(path: string, body: unknown, token: string): Promise<ApiEnvelope<T>> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token.trim()) {
    headers.Authorization = `Bearer ${token.trim()}`;
  }
  const res = await fetch(path, { method: 'POST', headers, body: JSON.stringify(body) });
  const raw = await res.text();
  let parsed: ApiEnvelope<T>;
  try {
    parsed = JSON.parse(raw) as ApiEnvelope<T>;
  } catch {
    throw new Error(`HTTP ${res.status}: ${raw}`);
  }
  if (res.status >= 400) {
    throw new Error(`HTTP ${res.status}: ${parsed.message || raw}`);
  }
  if (parsed.code !== 0 && parsed.code !== 2001) {
    throw new Error(`code=${parsed.code} ${parsed.message}`);
  }
  return parsed;
}

export async function checkUpdate(cfg: DeviceConfig) {
  return devicePost<CheckUpdateData>('/device/v1/check-update', {
    device_id: cfg.deviceId,
    group: cfg.group,
    product_model: cfg.productModel,
    hardware_version: cfg.hardwareVersion,
    current_version: cfg.currentVersion,
  }, cfg.apiToken);
}

export async function reportStatus(
  cfg: DeviceConfig,
  payload: { taskId: string; status: string; targetVersion: string },
) {
  return devicePost('/device/v1/report-status', {
    device_id: cfg.deviceId,
    task_id: payload.taskId,
    status: payload.status,
    source_version: cfg.currentVersion,
    target_version: payload.targetVersion,
  }, cfg.apiToken);
}

export const UPGRADE_STATUS_STEPS = ['pending', 'downloading', 'downloaded', 'upgrading', 'success'] as const;

export async function probeDownload(url: string): Promise<{ status: number; bytes: number }> {
  const res = await fetch(url, { method: 'GET' });
  const buf = await res.arrayBuffer();
  return { status: res.status, bytes: buf.byteLength };
}
