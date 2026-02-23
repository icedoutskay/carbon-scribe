import { SecurityEventType, SecuritySeverity } from '../constants/security-events.constants';

export interface SecurityEventContext {
  companyId?: string;
  userId?: string;
  ipAddress?: string;
  userAgent?: string;
  resource?: string;
  method?: string;
  details?: Record<string, unknown>;
  oldValue?: unknown;
  newValue?: unknown;
  status?: string;
  statusCode?: number;
}

export interface SecurityEvent extends SecurityEventContext {
  eventType: SecurityEventType;
  severity: SecuritySeverity;
}

