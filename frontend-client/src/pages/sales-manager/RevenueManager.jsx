import { useState } from "react";
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from "recharts";

const API_BASE = import.meta.env.VITE_API_URL;

function today() {
    return new Date().toISOString().slice(0, 10);
}
function daysAgo(n) {
    const d = new Date();
    d.setDate(d.getDate() - n);
    return d.toISOString().slice(0, 10);
}

export default function RevenueManager() {
    const [from, setFrom]     = useState(daysAgo(30));
    const [to, setTo]         = useState(today());
    const [data, setData]     = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError]   = useState(null);

    async function fetchRevenue() {
        if (!from || !to) { setError("Please select both dates."); return; }
        if (from > to)    { setError("Start date must be before end date."); return; }
        setLoading(true); setError(null); setData(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/admin/revenue?from=${from}&to=${to}`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            if (!res.ok) { const e = await res.json(); throw new Error(e.error || "Failed"); }
            setData(await res.json());
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    }

    const statCard = (label, value, color = "text-gray-900") => (
        <div className="bg-white border border-gray-200 rounded-xl px-5 py-4">
            <p className="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-1">{label}</p>
            <p className={`text-2xl font-bold ${color}`}>${Number(value).toFixed(2)}</p>
        </div>
    );

    return (
        <div className="max-w-4xl mx-auto">
            <div className="mb-6">
                <h1 className="text-2xl font-bold text-gray-900">Revenue</h1>
                <p className="text-sm text-gray-400 mt-0.5">Profit metrics across a custom date range</p>
            </div>

            {/* Date range controls */}
            <div className="bg-white border border-gray-200 rounded-xl px-5 py-4 mb-4 flex flex-wrap items-end gap-4">
                <div>
                    <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">From</label>
                    <input
                        type="date"
                        value={from}
                        onChange={e => setFrom(e.target.value)}
                        className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                </div>
                <div>
                    <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">To</label>
                    <input
                        type="date"
                        value={to}
                        onChange={e => setTo(e.target.value)}
                        className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                </div>
                <div className="flex gap-2">
                    {[7, 30, 90].map(n => (
                        <button
                            key={n}
                            onClick={() => { setFrom(daysAgo(n)); setTo(today()); }}
                            className="text-xs text-blue-600 border border-blue-200 hover:bg-blue-50 px-2.5 py-1.5 rounded-lg transition-colors"
                        >
                            Last {n}d
                        </button>
                    ))}
                </div>
                <button
                    onClick={fetchRevenue}
                    disabled={loading}
                    className="bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold px-5 py-2 rounded-lg disabled:opacity-50 transition-colors"
                >
                    {loading ? "Loading…" : "Show Report"}
                </button>
            </div>

            {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3 mb-4">{error}</div>
            )}

            {data && (
                <>
                    {/* Stat cards */}
                    <div className="grid grid-cols-3 gap-4 mb-4">
                        {statCard("Total Revenue", data.total_revenue, "text-blue-600")}
                        {statCard("Total Cost", data.total_cost, "text-gray-700")}
                        {statCard("Profit", data.profit, data.profit >= 0 ? "text-green-600" : "text-red-600")}
                    </div>

                    {/* Chart */}
                    {data.daily && data.daily.length > 0 ? (
                        <div className="bg-white border border-gray-200 rounded-xl p-5">
                            <h2 className="text-sm font-bold text-gray-700 mb-4">Daily Breakdown</h2>
                            <ResponsiveContainer width="100%" height={280}>
                                <LineChart data={data.daily} margin={{ top: 5, right: 20, left: 10, bottom: 5 }}>
                                    <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                                    <XAxis dataKey="date" tick={{ fontSize: 11 }} />
                                    <YAxis tick={{ fontSize: 11 }} tickFormatter={v => `$${v}`} />
                                    <Tooltip formatter={v => `$${Number(v).toFixed(2)}`} />
                                    <Legend />
                                    <Line type="monotone" dataKey="revenue" stroke="#2563eb" strokeWidth={2} dot={false} name="Revenue" />
                                    <Line type="monotone" dataKey="profit"  stroke="#16a34a" strokeWidth={2} dot={false} name="Profit" />
                                </LineChart>
                            </ResponsiveContainer>
                        </div>
                    ) : (
                        <div className="bg-white border border-gray-200 rounded-xl px-6 py-10 text-center">
                            <p className="text-3xl mb-2">📊</p>
                            <p className="text-gray-600 font-semibold">No orders in this period</p>
                            <p className="text-sm text-gray-400 mt-1">Try a wider date range.</p>
                        </div>
                    )}
                </>
            )}

            {!data && !loading && !error && (
                <div className="bg-white border border-gray-200 rounded-xl px-6 py-14 text-center">
                    <p className="text-3xl mb-2">📊</p>
                    <p className="text-gray-600 font-semibold">Select a date range and click Show Report</p>
                </div>
            )}
        </div>
    );
}