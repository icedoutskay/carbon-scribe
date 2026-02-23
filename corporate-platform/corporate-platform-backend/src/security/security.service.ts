import { Injectable } from '@nestjs/common';
import { PrismaService } from '../shared/database/prisma.service';
import { SecurityEvent } from './interfaces/security-event.interface';
import { SecurityEventType, SecuritySeverity, SecurityEvents } from './constants/security-events.constants';

@Injectable()
export class SecurityService {
  constructor(private readonly prisma: PrismaService) {}

  async isIpAllowed(companyId: string, ip: string, isAdminOverride: boolean): Promise<boolean> {
    if (isAdminOverride) {
      return true;
    }

    const entries = await (this.prisma as any).ipWhitelist.findMany({
      where: {
        companyId,
        isActive: true,
      },
    });

    if (entries.length === 0) {
      return true;
    }

    const ipNum = this.ipToNumber(ip);
    if (ipNum === null) {
      return false;
    }

    for (const entry of entries) {
      if (this.cidrContains(entry.cidr, ipNum)) {
        return true;
      }
    }

    return false;
  }

  private ipToNumber(ip: string): number | null {
    const parts = ip.split('.');
    if (parts.length !== 4) {
      return null;
    }
    let result = 0;
    for (const part of parts) {
      const n = Number(part);
      if (!Number.isInteger(n) || n < 0 || n > 255) {
        return null;
      }
      result = (result << 8) + n;
    }
    return result >>> 0;
  }

  private cidrContains(cidr: string, ipNum: number): boolean {
    const [base, prefixStr] = cidr.split('/');
    const prefix = Number(prefixStr ?? '32');
    const baseNum = this.ipToNumber(base);
    if (baseNum === null || !Number.isInteger(prefix) || prefix < 0 || prefix > 32) {
      return false;
    }
    const mask = prefix === 0 ? 0 : (~0 << (32 - prefix)) >>> 0;
    return (ipNum & mask) === (baseNum & mask);
  }

  async addWhitelistEntry(
    companyId: string,
    cidr: string,
    createdBy: string,
    description?: string,
  ): Promise<any> {
    const entry = await (this.prisma as any).ipWhitelist.create({
      data: {
        companyId,
        cidr,
        description,
        createdBy,
      },
    });

    await this.logEvent({
      eventType: SecurityEvents.IpWhitelistAdded,
      severity: 'info',
      companyId,
      userId: createdBy,
      details: { cidr, description, id: entry.id },
      status: 'success',
    });

    return entry;
  }

  async removeWhitelistEntry(companyId: string, id: string, userId: string): Promise<void> {
    const existing = await (this.prisma as any).ipWhitelist.findFirst({
      where: { id, companyId },
    });
    if (!existing) {
      return;
    }

    await (this.prisma as any).ipWhitelist.delete({
      where: { id },
    });

    await this.logEvent({
      eventType: SecurityEvents.IpWhitelistRemoved,
      severity: 'info',
      companyId,
      userId,
      details: { id: existing.id, cidr: existing.cidr },
      status: 'success',
    });
  }

  async listWhitelist(companyId: string): Promise<any[]> {
    return (this.prisma as any).ipWhitelist.findMany({
      where: { companyId },
      orderBy: { createdAt: 'desc' },
    });
  }

  async logEvent(event: SecurityEvent): Promise<any> {
    const status = event.status ?? 'success';
    const statusCode = event.statusCode ?? null;

    return (this.prisma as any).auditLog.create({
      data: {
        companyId: event.companyId ?? null,
        userId: event.userId ?? null,
        eventType: event.eventType,
        severity: event.severity,
        ipAddress: event.ipAddress ?? null,
        userAgent: event.userAgent ?? null,
        resource: event.resource ?? null,
        method: event.method ?? null,
        details: event.details ?? null,
        oldValue: event.oldValue ?? null,
        newValue: event.newValue ?? null,
        status,
        statusCode,
      },
    });
  }

  classifySeverity(eventType: SecurityEventType, status: string): SecuritySeverity {
    if (
      eventType === SecurityEvents.AuthLoginFailed ||
      eventType === SecurityEvents.AuthCrossTenantAttempt ||
      eventType === SecurityEvents.IpBlocked ||
      eventType === SecurityEvents.RateLimitExceeded
    ) {
      if (status === 'blocked' || status === 'failure') {
        return 'critical';
      }
      return 'warning';
    }

    if (eventType === SecurityEvents.SuspiciousPatternDetected) {
      return 'critical';
    }

    if (
      eventType === SecurityEvents.AuthPasswordChange ||
      eventType === SecurityEvents.AuthPasswordReset ||
      eventType === SecurityEvents.UserRoleChanged ||
      eventType === SecurityEvents.SettingsChanged
    ) {
      return 'warning';
    }

    return 'info';
  }

  async queryAuditLogs(
    companyId: string,
    params: {
      from?: Date;
      to?: Date;
      userId?: string;
      eventType?: string;
      severity?: string;
      limit?: number;
    },
  ): Promise<any[]> {
    const where: any = {
      companyId,
    };

    if (params.userId) {
      where.userId = params.userId;
    }

    if (params.eventType) {
      where.eventType = params.eventType;
    }

    if (params.severity) {
      where.severity = params.severity;
    }

    if (params.from || params.to) {
      where.timestamp = {};
      if (params.from) {
        where.timestamp.gte = params.from;
      }
      if (params.to) {
        where.timestamp.lte = params.to;
      }
    }

    const take = params.limit && params.limit > 0 && params.limit <= 1000 ? params.limit : 100;

    return (this.prisma as any).auditLog.findMany({
      where,
      orderBy: { timestamp: 'desc' },
      take,
    });
  }
}
