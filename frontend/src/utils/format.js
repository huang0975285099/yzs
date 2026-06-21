/**
 * Shared formatting utilities
 */

/**
 * Format datetime as MM-DD HH:mm (compact, used in tables)
 * @param {string|number|Date} t
 * @returns {string}
 */
export function formatTime(t) {
    if (!t) return '—'
    const d = new Date(t)
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    const hh = String(d.getHours()).padStart(2, '0')
    const min = String(d.getMinutes()).padStart(2, '0')
    return `${mm}-${dd} ${hh}:${min}`
}

/**
 * Format datetime using locale string (verbose, used in detail views)
 * @param {string|number|Date} t
 * @returns {string}
 */
export function formatTimeLocale(t) {
    if (!t) return '—'
    return new Date(t).toLocaleString('zh-CN')
}

/**
 * Format duration in seconds to Chinese readable string
 * @param {number} seconds
 * @returns {string}
 */
export function formatDuration(seconds) {
    if (!seconds) return '0秒'
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return m === 0 ? `${s}秒` : `${m}分${s}秒`
}

/**
 * Format video duration field (NULL / 0 / 1 / >1 seconds)
 * null  → '—'      未尝试
 * 0     → '无视频'
 * 1     → '解析失败'
 * >1    → '42秒' / '1分30秒'
 * @param {number|null} v
 * @returns {string}
 */
export function formatVideoDuration(v) {
    if (v === null || v === undefined) return '—'
    if (v === 0) return '无视频'
    if (v === 1) return '解析失败'
    const m = Math.floor(v / 60)
    const s = v % 60
    return m === 0 ? `${s}秒` : `${m}分${s}秒`
}

/**
 * Parse goods JSON string to array
 * @param {string} json
 * @returns {Array}
 */
export function parseGoods(json) {
    if (!json) return []
    try {
        return JSON.parse(json) || []
    } catch {
        return []
    }
}

/**
 * Generate a short goods summary string from JSON
 * e.g. "可乐×2、薯片×1  ¥15.00"
 * @param {string} json
 * @returns {string}
 */
export function goodsSummary(json) {
    if (!json) return ''
    try {
        const goods = JSON.parse(json)
        if (!Array.isArray(goods) || goods.length === 0) return '无消费'
        const list = goods.map((g) => `${g.goodsName}×${g.goodsCount}`).join('、')
        const total = goods.reduce((s, g) => s + (g.goodsPrice || 0) * (g.goodsCount || 0), 0)
        return `${list}  ¥${total.toFixed(2)}`
    } catch {
        return ''
    }
}
