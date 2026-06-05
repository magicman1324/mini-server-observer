import { useState, useEffect, useCallback } from 'react'
import { Plus, Edit, Trash2, ChevronLeft, ChevronRight } from 'lucide-react'
import { api } from '../api/client'
import type { Rule } from '../api/types'

const severityBadge: Record<string, string> = {
  critical: 'bg-red-500/10 text-red-400 border-red-500/30',
  warning: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30',
  info: 'bg-blue-500/10 text-blue-400 border-blue-500/30',
}

const operators = ['>', '>=', '<', '<=', '==', '!=']
const severities = ['critical', 'warning', 'info']

interface RuleForm {
  name: string
  description: string
  metric: string
  operator: string
  threshold: number
  duration_sec: number
  severity: string
  enabled: boolean
}

const emptyForm: RuleForm = {
  name: '', description: '', metric: '', operator: '>',
  threshold: 0, duration_sec: 30, severity: 'warning', enabled: true,
}

export default function Rules() {
  const [rules, setRules] = useState<Rule[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [showModal, setShowModal] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState<RuleForm>(emptyForm)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const pageSize = 10

  const fetchRules = useCallback(async (p: number) => {
    setLoading(true)
    try {
      const res = await api.getRules(p, pageSize)
      setRules(res.data as Rule[])
      setTotal(res.total)
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }, [])

  useEffect(() => { fetchRules(page) }, [page, fetchRules])

  const openCreate = () => {
    setEditingId(null)
    setForm(emptyForm)
    setError('')
    setShowModal(true)
  }

  const openEdit = (rule: Rule) => {
    setEditingId(rule.id)
    setForm({
      name: rule.name,
      description: rule.description,
      metric: rule.metric,
      operator: rule.operator,
      threshold: rule.threshold,
      duration_sec: rule.duration_sec,
      severity: rule.severity,
      enabled: rule.enabled,
    })
    setError('')
    setShowModal(true)
  }

  const handleSave = async () => {
    if (!form.name || !form.metric) {
      setError('名称和指标为必填项')
      return
    }
    setSaving(true)
    setError('')
    try {
      if (editingId) {
        await api.updateRule(editingId, form)
      } else {
        await api.createRule(form)
      }
      setShowModal(false)
      fetchRules(page)
    } catch {
      setError('保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('确定要删除这条规则吗？')) return
    try {
      await api.deleteRule(id)
      fetchRules(page)
    } catch { /* ignore */ }
  }

  const handleToggle = async (rule: Rule) => {
    try {
      await api.updateRule(rule.id, { ...rule, enabled: !rule.enabled })
      fetchRules(page)
    } catch { /* ignore */ }
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize))

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-semibold">规则管理</h2>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 px-4 py-2 bg-cyan-600 hover:bg-cyan-500 rounded-lg text-sm font-medium transition-colors"
        >
          <Plus size={16} /> 新建规则
        </button>
      </div>

      <div className="bg-gray-900 rounded-xl border border-gray-800 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-800/50 text-gray-400">
            <tr>
              <th className="text-left p-3 pl-4">名称</th>
              <th className="text-left p-3">指标</th>
              <th className="text-left p-3">条件</th>
              <th className="text-left p-3">持续</th>
              <th className="text-left p-3">级别</th>
              <th className="text-left p-3">状态</th>
              <th className="text-right p-3 pr-4">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800">
            {loading ? (
              <tr><td colSpan={7} className="p-6 text-center text-gray-500">加载中...</td></tr>
            ) : rules.length === 0 ? (
              <tr><td colSpan={7} className="p-6 text-center text-gray-500">暂无规则</td></tr>
            ) : (
              rules.map((rule) => (
                <tr key={rule.id} className="hover:bg-gray-800/30 transition-colors">
                  <td className="p-3 pl-4 font-medium">{rule.name}</td>
                  <td className="p-3 text-gray-400 font-mono text-xs">{rule.metric}</td>
                  <td className="p-3 font-mono">{rule.operator} {rule.threshold}</td>
                  <td className="p-3 text-gray-400">{rule.duration_sec}s</td>
                  <td className="p-3">
                    <span className={`text-xs px-2 py-0.5 rounded border ${severityBadge[rule.severity] || ''}`}>
                      {rule.severity}
                    </span>
                  </td>
                  <td className="p-3">
                    <button
                      onClick={() => handleToggle(rule)}
                      className={`text-xs px-2 py-0.5 rounded border transition-colors ${
                        rule.enabled
                          ? 'bg-green-500/10 text-green-400 border-green-500/30'
                          : 'bg-gray-500/10 text-gray-500 border-gray-500/30'
                      }`}
                    >
                      {rule.enabled ? '启用' : '禁用'}
                    </button>
                  </td>
                  <td className="p-3 pr-4">
                    <div className="flex items-center justify-end gap-2">
                      <button onClick={() => openEdit(rule)} className="p-1.5 hover:bg-gray-700 rounded transition-colors" title="编辑">
                        <Edit size={14} />
                      </button>
                      <button onClick={() => handleDelete(rule.id)} className="p-1.5 hover:bg-red-500/10 text-red-400 rounded transition-colors" title="删除">
                        <Trash2 size={14} />
                      </button>
                    </div>
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

      {showModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-900 rounded-xl border border-gray-800 w-full max-w-md p-6 space-y-4">
            <h3 className="text-lg font-semibold">{editingId ? '编辑规则' : '新建规则'}</h3>
            <div className="space-y-3">
              <div>
                <label className="block text-sm text-gray-400 mb-1">名称 *</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none focus:border-cyan-500"
                />
              </div>
              <div>
                <label className="block text-sm text-gray-400 mb-1">描述</label>
                <input
                  type="text"
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none focus:border-cyan-500"
                />
              </div>
              <div>
                <label className="block text-sm text-gray-400 mb-1">指标名 *</label>
                <input
                  type="text"
                  value={form.metric}
                  onChange={(e) => setForm({ ...form, metric: e.target.value })}
                  className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg font-mono text-sm focus:outline-none focus:border-cyan-500"
                  placeholder="cpu_usage_percent"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm text-gray-400 mb-1">操作符</label>
                  <select
                    value={form.operator}
                    onChange={(e) => setForm({ ...form, operator: e.target.value })}
                    className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none focus:border-cyan-500"
                  >
                    {operators.map((op) => <option key={op} value={op}>{op}</option>)}
                  </select>
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-1">阈值</label>
                  <input
                    type="number"
                    value={form.threshold}
                    onChange={(e) => setForm({ ...form, threshold: Number(e.target.value) })}
                    className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none focus:border-cyan-500"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm text-gray-400 mb-1">持续时间 (秒)</label>
                  <input
                    type="number"
                    value={form.duration_sec}
                    onChange={(e) => setForm({ ...form, duration_sec: Number(e.target.value) })}
                    className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none focus:border-cyan-500"
                  />
                </div>
                <div>
                  <label className="block text-sm text-gray-400 mb-1">严重级别</label>
                  <select
                    value={form.severity}
                    onChange={(e) => setForm({ ...form, severity: e.target.value })}
                    className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg focus:outline-none focus:border-cyan-500"
                  >
                    {severities.map((s) => <option key={s} value={s}>{s}</option>)}
                  </select>
                </div>
              </div>
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={form.enabled}
                  onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
                  className="rounded"
                />
                <span className="text-gray-400">启用</span>
              </label>
            </div>
            {error && <p className="text-red-400 text-sm">{error}</p>}
            <div className="flex gap-3 justify-end pt-2">
              <button
                onClick={() => setShowModal(false)}
                className="px-4 py-2 rounded-lg text-sm hover:bg-gray-800 transition-colors"
              >
                取消
              </button>
              <button
                onClick={handleSave}
                disabled={saving}
                className="px-4 py-2 bg-cyan-600 hover:bg-cyan-500 disabled:opacity-50 rounded-lg text-sm font-medium transition-colors"
              >
                {saving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
