import { useState, useEffect } from 'react'
import { AlertTriangle, CheckCircle, ChevronLeft, ChevronRight } from 'lucide-react'
import { api } from '../api/client'
import type { Alert } from '../api/types'

const severityBadge: Record<string, string> = {
  critical: 'bg-red-500/10 text-red-400 border-red-500/30',
  warning: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30',
  info: 'bg-blue-500/10 text-blue-400 border-blue-500/30',
}

const statusIcon: Record<string, React.ReactNode> = {
  firing: <AlertTriangle size={14} className="text-red-400" />,
  resolved: <CheckCircle size={14} className="text-green-400" />,
}

export default function Alerts() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState('')
  const [severity, setSeverity] = useState('')
  const [loading, setLoading] = useState(true)
  const pageSize = 15

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    api.getAlerts(page, pageSize, status, severity).then((res) => {
      if (cancelled) return
      setAlerts(res.data as Alert[])
      setTotal(res.total)
      setLoading(false)
    }).catch(() => {
      if (!cancelled) setLoading(false)
    })
    return () => { cancelled = true }
  }, [page, status, severity])

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  const formatTime = (ts: string) => {
    const d = new Date(ts)
    return d.toLocaleString('zh-CN', { hour12: false })
  }

  return (
    <div>
      <h2 className="text-xl font-semibold mb-6">告警列表</h2>

      <div className="flex gap-3 mb-4">
        <select
          value={status}
          onChange={(e) => { setStatus(e.target.value); setPage(1) }}
          className="px-3 py-2 bg-gray-900 border border-gray-700 rounded-lg text-sm focus:outline-none focus:border-cyan-500"
        >
          <option value="">全部状态</option>
          <option value="firing">触发中</option>
          <option value="resolved">已恢复</option>
        </select>
        <select
          value={severity}
          onChange={(e) => { setSeverity(e.target.value); setPage(1) }}
          className="px-3 py-2 bg-gray-900 border border-gray-700 rounded-lg text-sm focus:outline-none focus:border-cyan-500"
        >
          <option value="">全部级别</option>
          <option value="critical">critical</option>
          <option value="warning">warning</option>
          <option value="info">info</option>
        </select>
      </div>

      <div className="bg-gray-900 rounded-xl border border-gray-800 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-800/50 text-gray-400">
            <tr>
              <th className="text-left p-3 pl-4 w-6"></th>
              <th className="text-left p-3">级别</th>
              <th className="text-left p-3">规则</th>
              <th className="text-left p-3">主机</th>
              <th className="text-left p-3">当前值</th>
              <th className="text-left p-3">阈值</th>
              <th className="text-left p-3">触发时间</th>
              <th className="text-left p-3">状态</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800">
            {loading ? (
              <tr><td colSpan={8} className="p-6 text-center text-gray-500">加载中...</td></tr>
            ) : alerts.length === 0 ? (
              <tr><td colSpan={8} className="p-6 text-center text-gray-500">暂无告警</td></tr>
            ) : (
              alerts.map((alert) => (
                <tr key={alert.id} className="hover:bg-gray-800/30 transition-colors">
                  <td className="p-3 pl-4">{statusIcon[alert.status]}</td>
                  <td className="p-3">
                    <span className={`text-xs px-2 py-0.5 rounded border ${severityBadge[alert.severity] || ''}`}>
                      {alert.severity}
                    </span>
                  </td>
                  <td className="p-3 font-medium">{alert.rule_name}</td>
                  <td className="p-3 text-gray-400 font-mono text-xs">{alert.host}</td>
                  <td className="p-3 font-mono">{alert.value.toFixed(1)}</td>
                  <td className="p-3 text-gray-400">{alert.threshold}</td>
                  <td className="p-3 text-gray-400 text-xs">{formatTime(alert.fired_at)}</td>
                  <td className="p-3">
                    <span className={`text-xs px-2 py-0.5 rounded border ${
                      alert.status === 'firing'
                        ? 'bg-red-500/10 text-red-400 border-red-500/30'
                        : 'bg-green-500/10 text-green-400 border-green-500/30'
                    }`}>
                      {alert.status === 'firing' ? '触发中' : '已恢复'}
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
        {totalPages > 1 && (
          <div className="flex items-center justify-between p-3 border-t border-gray-800 text-sm text-gray-400">
            <span>共 {total} 条</span>
            <div className="flex items-center gap-2">
              <button
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
                className="p-1 hover:bg-gray-700 rounded disabled:opacity-30 transition-colors"
              >
                <ChevronLeft size={16} />
              </button>
              <span>{page} / {totalPages}</span>
              <button
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className="p-1 hover:bg-gray-700 rounded disabled:opacity-30 transition-colors"
              >
                <ChevronRight size={16} />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
