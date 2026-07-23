export function canManageDeviceSecrets(roles: string[]): boolean {
  return roles.includes('admin') || roles.includes('secret_admin');
}

export function canManageUpgradePolicy(roles: string[]): boolean {
  return roles.includes('admin');
}
