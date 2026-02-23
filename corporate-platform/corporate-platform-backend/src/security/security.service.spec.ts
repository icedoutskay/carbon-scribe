import { Test, TestingModule } from '@nestjs/testing';
import { SecurityService } from './security.service';

describe('SecurityService', () => {
  let service: SecurityService;
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

    const module: TestingModule = await Test.createTestingModule({
      providers: [
        SecurityService,
        {
          provide: 'PrismaService',
          useValue: prisma,
        } as any,
      ],
    }).compile();

    service = module.get<SecurityService>(SecurityService);
  });

  it('should be defined', () => {
    expect(service).toBeDefined();
  });

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
  });
});

