import { MiddlewareConsumer, Module, NestModule } from '@nestjs/common';
import { SecurityService } from './security.service';
import { AuditLogMiddleware } from './middleware/audit-log.middleware';
import { DatabaseModule } from '../shared/database/database.module';
import { IpWhitelistGuard } from './guards/ip-whitelist.guard';
import { APP_GUARD } from '@nestjs/core';
import { SecurityController } from './security.controller';

@Module({
  imports: [DatabaseModule],
  providers: [
    SecurityService,
    {
      provide: APP_GUARD,
      useClass: IpWhitelistGuard,
    },
  ],
  controllers: [SecurityController],
  exports: [SecurityService],
})
export class SecurityModule implements NestModule {
  configure(consumer: MiddlewareConsumer) {
    consumer.apply(AuditLogMiddleware).forRoutes('*');
  }
}

