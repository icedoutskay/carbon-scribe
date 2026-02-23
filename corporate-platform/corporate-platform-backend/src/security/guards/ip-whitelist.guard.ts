import {
  CanActivate,
  ExecutionContext,
  Injectable,
  ForbiddenException,
} from '@nestjs/common';
import { Request } from 'express';
import { SecurityService } from '../security.service';
import { JwtPayload } from '../../auth/interfaces/jwt-payload.interface';
import { Reflector } from '@nestjs/core';

export const IP_WHITELIST_SKIP_KEY = 'ipWhitelistSkip';

@Injectable()
export class IpWhitelistGuard implements CanActivate {
  constructor(
    private readonly securityService: SecurityService,
    private readonly reflector: Reflector,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const skip = this.reflector.getAllAndOverride<boolean>(IP_WHITELIST_SKIP_KEY, [
      context.getHandler(),
      context.getClass(),
    ]);

    if (skip) {
      return true;
    }

    const http = context.switchToHttp();
    const req = http.getRequest<Request>();
    const user = req.user as JwtPayload | undefined;

    if (!user || !user.companyId) {
      return true;
    }

    const ip =
      (req.headers['x-forwarded-for'] as string | undefined)?.split(',')[0].trim() ||
      req.socket.remoteAddress ||
      '';

    const isAdminOverride = user.role === 'admin';

    const allowed = await this.securityService.isIpAllowed(user.companyId, ip, isAdminOverride);
    if (!allowed) {
      await this.securityService.logEvent({
        eventType: 'ip.blocked',
        severity: 'warning',
        companyId: user.companyId,
        userId: user.sub,
        ipAddress: ip,
        resource: req.originalUrl,
        method: req.method,
        status: 'blocked',
        statusCode: 403,
      });

      throw new ForbiddenException('Access from this IP is not allowed');
    }

    return true;
  }
}

