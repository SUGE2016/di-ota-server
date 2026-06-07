/** 解析 JWT payload（仅用于展示，鉴权仍由服务端校验） */
export function parseJwtSub(token: string): string | null {
  try {
    const payload = token.split('.')[1];
    if (!payload) return null;
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/');
    const json = JSON.parse(atob(normalized)) as { sub?: unknown };
    const sub = typeof json.sub === 'string' ? json.sub.trim() : '';
    return sub || null;
  } catch {
    return null;
  }
}
