export const SecurityEvents = {
  AuthLoginSuccess: 'auth.login.success',
  AuthLoginFailed: 'auth.login.failed',
  AuthLogout: 'auth.logout',
  AuthPasswordChange: 'auth.password.change',
  AuthPasswordReset: 'auth.password.reset',
  AuthPermissionDenied: 'auth.permission.denied',
  AuthRoleChanged: 'auth.role.changed',
  AuthCrossTenantAttempt: 'auth.cross-tenant.attempt',
  ApiKeyCreated: 'apikey.created',
  ApiKeyRevoked: 'apikey.revoked',
  ApiKeyRotated: 'apikey.rotated',
  ApiKeyRateLimitExceeded: 'apikey.rate-limit.exceeded',
  IpBlocked: 'ip.blocked',
  IpWhitelistAdded: 'ip.whitelist.added',
  IpWhitelistRemoved: 'ip.whitelist.removed',
  UserCreated: 'user.created',
  UserUpdated: 'user.updated',
  UserDeleted: 'user.deleted',
  UserRoleChanged: 'user.role.changed',
  CreditRetired: 'credit.retired',
  ReportExported: 'report.exported',
  SettingsChanged: 'settings.changed',
  RateLimitExceeded: 'rate-limit.exceeded',
  SuspiciousPatternDetected: 'suspicious.pattern.detected',
} as const;

export type SecurityEventType = (typeof SecurityEvents)[keyof typeof SecurityEvents];

export type SecuritySeverity = 'info' | 'warning' | 'critical';

