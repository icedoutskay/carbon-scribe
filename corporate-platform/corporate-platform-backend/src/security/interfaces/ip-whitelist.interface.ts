export interface IpWhitelistEntry {
  id: string;
  companyId: string;
  cidr: string;
  description?: string | null;
  isActive: boolean;
  createdBy: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface IpWhitelistCheckResult {
  allowed: boolean;
  reason?: string;
}

