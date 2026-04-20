export type CustomDomain = {
  id: string;
  name: string;
  type: 'system' | 'custom';
  status: 'pending_dns' | 'dns_verified' | 'active' | 'failed' | 'removing';
  target_subdomain: string | null;
  dns_provider: string | null;
  verification_token: string | null;
  created_at: string;
  updated_at: string;
};

export type DNSInstruction = {
  record_type: 'CNAME' | 'A' | 'AAAA' | 'TXT';
  name: string;
  value: string;
  description: string;
};

export type AddCustomDomainResponse = {
  status: string;
  message: string;
  data: CustomDomain;
  instructions: DNSInstruction[];
  dns_provider: string;
};

export type DNSCheckResponse = {
  status: string;
  message: string;
  verified: boolean;
  dns_status: string;
};
