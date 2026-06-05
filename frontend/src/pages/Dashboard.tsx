import { useState, useEffect } from 'react'
import { Monitor, AlertTriangle, Bell, ShieldCheck } from 'lucide-react'
import { api } from '../api/client'

interface Stats {
  hosts: number
  alertsFiring: number
  totalAlerts: number
  rulesEnabled: number
}

export default function Dashboard() {
  const [stats, setStats] = useState<Stats>({ hosts: 0, alertsFiring: 0, totalAlerts: 0, rulesEnabled: 0 })
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchStats() {
      try {
        const [hostsRes, alertStatsRes, rulesRes] = await Promise.all([
          api.getHosts(),
          api.getAlertStats(),
          api.getRules(1, 1),
        ])
        setStats({
          hosts: hostsRes.count,
          alertsFiring: alertStatsRes.firing,
          totalAlerts: alertStatsRes.total,
          rulesEnabled: rulesRes.total,
        })
      } catch {
        // use defaults
      } finally {
        setLoading(false)
      }
    }
    fetchStats()
  }, [])

  const cards = [
    { label: '监控主机', value: stats.hosts, icon: Monitor, color: 'text-cyan-400', bg: 'bg-cyan-500/10' },
    { label: '触发告警', value: stats.alertsFiring, icon: AlertTriangle, color: 'text-red-400', bg: 'bg-red-500/10' },
    { label: '告警总数', value: stats.totalAlerts, icon: Bell, color: 'text-yellow-400', bg: 'bg-yellow-500/10' },
    { label: '启用规则', value: stats.rulesEnabled, icon: ShieldCheck, color: 'text-green-400', bg: 'bg-green-500/10' },
  ]

  return (
    <div>
      <h2 className="text-xl font-semibold mb-6">系统概览</h2>
      {loading ? (
        <p className="text-gray-500">加载中...</p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {cards.map(({ label, value, icon: Icon, color, bg }) => (
            <div key={label} className="bg-gray-900 rounded-xl p-5 border border-gray-800">
              <div className="flex items-center justify-between mb-3">
                <span className="text-sm text-gray-500">{label}</span>
                <div className={`p-2 rounded-lg ${bg}`}>
                  <Icon size={18} className={color} />
                </div>
              </div>
              <p className={`text-2xl font-bold ${color}`}>{value}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
