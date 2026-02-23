import { Test, TestingModule } from '@nestjs/testing';
import { SecurityService } from './security.service';
import { PrismaService } from '../shared/database/prisma.service';

describe('SecurityService', () => {
  let service: SecurityService;
<<<<<<< HEAD
  let prisma: any;

  beforeEach(async () => {
    prisma = {
      ipWhitelist: {
        findMany: jest.fn(),
        create: jest.fn(),
        findFirst: jest.fn(),
        delete: jest.fn(),
      },
      auditLog: {
        create: jest.fn(),
        findMany: jest.fn(),
      },
    };

=======

  const prismaMock = {
    ipWhitelist: {
      findMany: jest.fn(),
      create: jest.fn(),
      findUnique: jest.fn(),
      delete: jest.fn(),
    },
    auditLog: {
      create: jest.fn(),
      findMany: jest.fn(),
      count: jest.fn(),
      deleteMany: jest.fn(),
    },
  } as any;

  beforeEach(async () => {
>>>>>>> fc2abbd (fixed: previous commit)
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        SecurityService,
        {
          provide: PrismaService,
<<<<<<< HEAD
          useValue: prisma,
=======
          useValue: prismaMock,
>>>>>>> fc2abbd (fixed: previous commit)
        },
      ],
    }).compile();

    service = module.get<SecurityService>(SecurityService);
<<<<<<< HEAD
=======

    prismaMock.ipWhitelist.findMany.mockReset();
    prismaMock.ipWhitelist.create.mockReset();
    prismaMock.ipWhitelist.findUnique.mockReset();
    prismaMock.ipWhitelist.delete.mockReset();
    prismaMock.auditLog.create.mockReset();
    prismaMock.auditLog.findMany.mockReset();
    prismaMock.auditLog.count.mockReset();
    prismaMock.auditLog.deleteMany.mockReset();
>>>>>>> fc2abbd (fixed: previous commit)
  });

  it('should be defined', () => {
    expect(service).toBeDefined();
  });

<<<<<<< HEAD
  it('allows ip when no entries', async () => {
    prisma.ipWhitelist.findMany.mockResolvedValue([]);
    const allowed = await service.isIpAllowed('c1', '192.168.0.10', false);
    expect(allowed).toBe(true);
  });

  it('matches cidr correctly', async () => {
    prisma.ipWhitelist.findMany.mockResolvedValue([
      { cidr: '192.168.0.0/24', companyId: 'c1', isActive: true },
    ]);
    const allowed = await service.isIpAllowed('c1', '192.168.0.10', false);
    expect(allowed).toBe(true);
  });

  it('blocks non matching ip', async () => {
    prisma.ipWhitelist.findMany.mockResolvedValue([
      { cidr: '10.0.0.0/8', companyId: 'c1', isActive: true },
    ]);
    const allowed = await service.isIpAllowed('c1', '192.168.0.10', false);
    expect(allowed).toBe(false);
=======
  it('validates IPv4 addresses correctly', () => {
    expect((service as any).isValidIpv4('192.168.1.1')).toBe(true);
    expect((service as any).isValidIpv4('0.0.0.0')).toBe(true);
    expect((service as any).isValidIpv4('255.255.255.255')).toBe(true);
    expect((service as any).isValidIpv4('999.0.0.1')).toBe(false);
    expect((service as any).isValidIpv4('abc.def.ghi.jkl')).toBe(false);
  });

  it('checks CIDR membership correctly', () => {
    expect((service as any).isIpInCidr('192.168.1.10', '192.168.1.0/24')).toBe(
      true,
    );
    expect((service as any).isIpInCidr('192.168.2.10', '192.168.1.0/24')).toBe(
      false,
    );
    expect((service as any).isIpInCidr('10.0.5.1', '10.0.0.0/8')).toBe(true);
  });

  it('logs audit events with required fields', async () => {
    prismaMock.auditLog.create.mockResolvedValue({});
    prismaMock.auditLog.deleteMany.mockResolvedValue({});

    await service.logEvent({
      eventType: 'auth.login.success' as any,
      status: 'success',
      statusCode: 200,
      companyId: 'c1',
      userId: 'u1',
      ipAddress: '127.0.0.1',
      userAgent: 'jest',
      resource: '/api/v1/auth/login',
      method: 'POST',
    });

    expect(prismaMock.auditLog.create).toHaveBeenCalledTimes(1);
    const call = prismaMock.auditLog.create.mock.calls[0][0];
    expect(call.data.companyId).toBe('c1');
    expect(call.data.userId).toBe('u1');
    expect(call.data.eventType).toBe('auth.login.success');
    expect(call.data.status).toBe('success');
>>>>>>> fc2abbd (fixed: previous commit)
  });
});
