import { Injectable, NestMiddleware } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';
import { SecurityService } from '../security.service';
import { JwtPayload } from '../../auth/interfaces/jwt-payload.interface';

@Injectable()
export class AuditLogMiddleware implements NestMiddleware {
  constructor(private readonly securityService: SecurityService) {}

  use(req: Request, res: Response, next: NextFunction): void {
    const start = Date.now();

    const done = () => {
      res.removeListener('finish', done);
      res.removeListener('close', done);

      const user = req.user as JwtPayload | undefined;

      const statusCode = res.statusCode;
      const status = statusCode >= 200 && statusCode < 400 ? 'success' : 'failure';

      this.securityService.logEvent({
        eventType: 'request',
        severity: 'info',
        companyId: user?.companyId,
        userId: user?.sub,
        ipAddress:
          (req.headers['x-forwarded-for'] as string | undefined)?.split(',')[0].trim() ||
          req.socket.remoteAddress ||
          undefined,
        userAgent: req.headers['user-agent'] as string | undefined,
        resource: req.originalUrl,
        method: req.method,
        status,
        statusCode,
        details: {
          durationMs: Date.now() - start,
        },
      } as any);
    };

    res.on('finish', done);
    res.on('close', done);

    next();
  }
}

