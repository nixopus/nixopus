export interface Domain {
  id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface RandomSubdomainResponse {
  subdomain: string;
  domain: string;
}

export interface CreateDomainRequest {
  name: string;
  organization_id: string;
}

export interface UpdateDomainRequest {
  name: string;
  id: string;
}

export interface DeleteDomainRequest {
  id: string;
}
