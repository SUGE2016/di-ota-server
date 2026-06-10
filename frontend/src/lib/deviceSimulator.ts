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
  current_version?: string;
};

class DeviceApiError extends Error {
  code: number;
  data?: Record<string, unknown>;

  constructor(code: number, message: string, data?: Record<string, unknown>) {
    super(message);
    this.code = code;
    this.data = data;
  }
}

function formatDeviceApiError(parsed: ApiEnvelope): DeviceApiError {
  const data = parsed.data as Record<string, unknown> | undefined;
  if (parsed.code === 2005 && data?.previous && data?.current) {
    return new DeviceApiError(
      parsed.code,
      `该任务当前状态为 ${data.previous}，无法回退上报 ${data.current}。请新建发布任务，或更换 device_id 后重试。`,
      data,
    );
  }
  return new DeviceApiError(parsed.code, `code=${parsed.code} ${parsed.message}`, data);
}

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
    throw formatDeviceApiError(parsed);
  }
  return parsed;
}

const deviceApi = (path: string) =>
  `${import.meta.env.BASE_URL}device/v1${path}`.replace(/\/{2,}/g, '/');

export async function checkUpdate(cfg: DeviceConfig) {
  return devicePost<CheckUpdateData>(deviceApi('/check-update'), {
    device_id: cfg.deviceId,
    group: cfg.group,
    product_model: cfg.productModel,
    hardware_version: cfg.hardwareVersion,
    current_version: cfg.currentVersion,
  }, cfg.apiToken);
}

export async function reportStatus(
  cfg: DeviceConfig,
  payload: { taskId: string; status: string; targetVersion: string; sourceVersion?: string },
) {
  return devicePost(deviceApi('/report-status'), {
    device_id: cfg.deviceId,
    task_id: payload.taskId,
    status: payload.status,
    source_version: payload.sourceVersion ?? cfg.currentVersion,
    target_version: payload.targetVersion,
  }, cfg.apiToken);
}

export const UPGRADE_STATUS_STEPS = ['pending', 'downloading', 'downloaded', 'upgrading', 'success'] as const;

export async function probeDownload(url: string): Promise<{ status: number; bytes: number }> {
  const res = await fetch(url, { method: 'GET' });
  const buf = await res.arrayBuffer();
  return { status: res.status, bytes: buf.byteLength };
}

export function describeNoUpdate(data: CheckUpdateData | undefined): string {
  const reason = data?.reason || 'unknown';
  const version = data?.current_version;
  if (reason === 'already_latest') {
    return version
      ? `服务端认定当前版本 ${version}，已达目标版本。重复联调请发布更高版本包并创建新 Running 任务。`
      : '服务端认定设备已是最新版本。重复联调请发布更高版本包并创建新 Running 任务。';
  }
  if (reason === 'no_running_task') {
    return '无 Running 任务或未命中任务快照，请先在管理台创建并启动发布任务。';
  }
  return `无可用升级（${reason}）`;
}
