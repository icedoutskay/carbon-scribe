import { Controller, Get, Post, Delete, Body, Param, Query, Req, UseGuards } from '@nestjs/common';
import { SecurityService } from './security.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { CurrentUser } from '../auth/decorators/current-user.decorator';
import { JwtPayload } from '../auth/interfaces/jwt-payload.interface';
import { Request } from 'express';

class CreateWhitelistDto {
  cidr: string;
  description?: string;
}

class QueryAuditLogsDto {
  from?: string;
  to?: string;
  userId?: string;
  eventType?: string;
  severity?: string;
  limit?: number;
}

@Controller('api/v1/security')
@UseGuards(JwtAuthGuard)
export class SecurityController {
  constructor(private readonly securityService: SecurityService) {}

  @Get('whitelist')
  async listWhitelist(@CurrentUser() user: JwtPayload) {
    return this.securityService.listWhitelist(user.companyId);
  }

  @Post('whitelist')
  async addWhitelist(@CurrentUser() user: JwtPayload, @Body() body: CreateWhitelistDto) {
    return this.securityService.addWhitelistEntry(
      user.companyId,
      body.cidr,
      user.sub,
      body.description,
    );
  }

  @Delete('whitelist/:id')
  async removeWhitelist(@CurrentUser() user: JwtPayload, @Param('id') id: string) {
    await this.securityService.removeWhitelistEntry(user.companyId, id, user.sub);
    return { success: true };
  }

  @Get('audit-logs')
  async getAuditLogs(
    @CurrentUser() user: JwtPayload,
    @Query() query: QueryAuditLogsDto,
  ): Promise<any[]> {
    const from = query.from ? new Date(query.from) : undefined;
    const to = query.to ? new Date(query.to) : undefined;
    return this.securityService.queryAuditLogs(user.companyId, {
      from,
      to,
      userId: query.userId,
      eventType: query.eventType,
      severity: query.severity,
      limit: query.limit,
    });
  }

  @Get('audit-logs/export')
  async exportAuditLogs(
    @CurrentUser() user: JwtPayload,
    @Query() query: QueryAuditLogsDto,
  ): Promise<{ items: any[] }> {
    const logs = await this.getAuditLogs(user, query);
    return { items: logs };
  }

  @Get('events/summary')
  async getSecuritySummary(@CurrentUser() user: JwtPayload, @Req() req: Request) {
    const now = new Date();
    const from = new Date(now.getTime() - 24 * 60 * 60 * 1000);

    const logs = await this.securityService.queryAuditLogs(user.companyId, {
      from,
      to: now,
      limit: 1000,
    });

    const bySeverity = new Map<string, number>();
    const byEventType = new Map<string, number>();

    for (const log of logs) {
      bySeverity.set(log.severity, (bySeverity.get(log.severity) || 0) + 1);
      byEventType.set(log.eventType, (byEventType.get(log.eventType) || 0) + 1);
    }

    return {
      since: from.toISOString(),
      until: now.toISOString(),
      totalEvents: logs.length,
      bySeverity: Object.fromEntries(bySeverity),
      topEventTypes: Array.from(byEventType.entries())
        .sort((a, b) => b[1] - a[1])
        .slice(0, 10)
        .map(([eventType, count]) => ({ eventType, count })),
    };
  }
}
