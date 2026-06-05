export interface MetricPoint {
  TS: string
  Name: string
  Value: number
  Host: string
  Region: string
}

export interface Rule {
  id: string
  name: string
  description: string
  metric: string
  operator: string
  threshold: number
  duration_sec: number
  severity: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface Alert {
  id: string
  rule_id: string
  rule_name: string
  host: string
  region: string
  severity: string
  metric: string
  value: number
  threshold: number
  message: string
  status: string
  fired_at: string
  resolved_at: string | null
}

export interface Host {
  host: string
  region: string
  last_seen: string
}

export interface Paginated<T> {
  data: T[]
  total: number
  page: number
  page_size: number
}

export interface DashboardStats {
  hosts: number
  alertsFiring: number
  totalAlerts: number
  rulesEnabled: number
}
