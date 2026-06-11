import { useState } from "react";

const API_BASE = import.meta.env.VITE_API_URL;

function today() { return new Date().toISOString().slice(0, 10); }
function daysAgo(n) { const d = new Date(); d.setDate(d.getDate() - n); return d.toISOString().slice(0, 10); }

// Fetch a PDF endpoint as a blob (Bearer auth) and trigger a browser download.
async function downloadPdf(url, filename) {
    const token = localStorage.getItem("token");
    const res = await fetch(`${API_BASE}${url}`, { headers: { Authorization: `Bearer ${token}` } });
    if (!res.ok) throw new Error("Failed to generate PDF");
    const blob = await res.blob();
    const objectUrl = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = objectUrl;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    window.URL.revokeObjectURL(objectUrl);
}

function OrderRow({ order }) {
    const [expanded, setExpanded] = useState(false);
    const [pdfErr, setPdfErr]     = useState(false);
    const date = new Date(order.created_at).toLocaleDateString("en-US", {
        year: "numeric", month: "short", day: "numeric",
    });

    async function handleRowPdf() {
        setPdfErr(false);
        try {
            await downloadPdf(`/api/orders/${order.delivery_id}/invoice`, `invoice_${order.delivery_id}.pdf`);
        } catch {
            setPdfErr(true);
        }
    }

    const STATUS_STYLES = {
        processing:   "bg-yellow-100 text-yellow-700",
        "in-transit": "bg-blue-100 text-blue-700",
        delivered:    "bg-green-100 text-green-700",
        cancelled:    "bg-red-100 text-red-700",
        returned:     "bg-gray-100 text-gray-600",
    };

    return (
        <>
            <tr className="border-t border-gray-100 hover:bg-gray-50 transition-colors">
                <td className="px-5 py-3 text-gray-700 font-medium">#{order.delivery_id}</td>
                <td className="px-4 py-3 text-gray-600">Customer #{order.customer_id}</td>
                <td className="px-4 py-3 text-gray-500 text-sm">{date}</td>
                <td className="px-4 py-3">
                    <span className={`text-xs font-semibold px-2.5 py-1 rounded-full ${STATUS_STYLES[order.status] || "bg-gray-100 text-gray-600"}`}>
                        {order.status}
                    </span>
                </td>
                <td className="px-4 py-3 text-right font-bold text-gray-900">${order.total_price.toFixed(2)}</td>
                <td className="px-5 py-3 text-right whitespace-nowrap">
                    <button
                        onClick={() => setExpanded(e => !e)}
                        className="text-xs text-blue-600 border border-blue-200 hover:bg-blue-50 px-2.5 py-1 rounded-lg transition-colors"
                    >
                        {expanded ? "Hide" : "View"} items
                    </button>
                    <button
                        onClick={handleRowPdf}
                        className="ml-2 text-xs text-gray-600 border border-gray-200 hover:bg-gray-50 px-2.5 py-1 rounded-lg transition-colors"
                        title="Download this invoice as PDF"
                    >
                        {pdfErr ? "Retry PDF" : "PDF"}
                    </button>
                </td>
            </tr>
            {expanded && order.items?.length > 0 && (
                <tr className="border-t border-gray-100 bg-gray-50">
                    <td colSpan="6" className="px-5 py-3">
                        <table className="w-full text-xs">
                            <thead>
                            <tr className="text-gray-400 uppercase tracking-wide">
                                <th className="text-left py-1 font-semibold">Product ID</th>
                                <th className="text-center py-1 font-semibold">Qty</th>
                                <th className="text-right py-1 font-semibold">Unit Price</th>
                                <th className="text-right py-1 font-semibold">Subtotal</th>
                            </tr>
                            </thead>
                            <tbody>
                            {order.items.map(item => (
                                <tr key={item.id} className="border-t border-gray-200">
                                    <td className="py-1.5 text-gray-600 font-mono">{item.product_id}</td>
                                    <td className="py-1.5 text-gray-600 text-center">{item.quantity}</td>
                                    <td className="py-1.5 text-gray-600 text-right">${item.price.toFixed(2)}</td>
                                    <td className="py-1.5 text-gray-800 font-semibold text-right">${(item.price * item.quantity).toFixed(2)}</td>
                                </tr>
                            ))}
                            </tbody>
                        </table>
                    </td>
                </tr>
            )}
        </>
    );
}

export default function InvoiceList() {
    const [from, setFrom]     = useState(daysAgo(30));
    const [to, setTo]         = useState(today());
    const [orders, setOrders] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError]   = useState(null);
    const [savingAll, setSavingAll] = useState(false);

    async function handleSaveAll() {
        setSavingAll(true);
        setError(null);
        try {
            await downloadPdf(`/api/admin/invoices?from=${from}&to=${to}&format=pdf`, `invoices_${from}_${to}.pdf`);
        } catch {
            setError("Failed to export invoices as PDF.");
        } finally {
            setSavingAll(false);
        }
    }

    async function fetchInvoices() {
        if (!from || !to) { setError("Please select both dates."); return; }
        if (from > to)    { setError("Start date must be before end date."); return; }
        setLoading(true); setError(null); setOrders(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/admin/invoices?from=${from}&to=${to}`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            if (!res.ok) { const e = await res.json(); throw new Error(e.error || "Failed"); }
            const data = await res.json();
            data.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
            setOrders(data);
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    }

    return (
        <div className="max-w-4xl mx-auto">
            <div className="mb-6">
                <h1 className="text-2xl font-bold text-gray-900">Invoices</h1>
                <p className="text-sm text-gray-400 mt-0.5">Browse all orders within a date range</p>
            </div>

            {/* Controls */}
            <div className="bg-white border border-gray-200 rounded-xl px-5 py-4 mb-4 flex flex-wrap items-end gap-4">
                <div>
                    <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">From</label>
                    <input type="date" value={from} onChange={e => setFrom(e.target.value)}
                           className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div>
                    <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">To</label>
                    <input type="date" value={to} onChange={e => setTo(e.target.value)}
                           className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div className="flex gap-2">
                    {[7, 30, 90].map(n => (
                        <button key={n} onClick={() => { setFrom(daysAgo(n)); setTo(today()); }}
                                className="text-xs text-blue-600 border border-blue-200 hover:bg-blue-50 px-2.5 py-1.5 rounded-lg transition-colors">
                            Last {n}d
                        </button>
                    ))}
                </div>
                <button onClick={fetchInvoices} disabled={loading}
                        className="bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold px-5 py-2 rounded-lg disabled:opacity-50 transition-colors">
                    {loading ? "Loading…" : "Search"}
                </button>
            </div>

            {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3 mb-4">{error}</div>
            )}

            {orders !== null && (
                orders.length === 0 ? (
                    <div className="bg-white border border-gray-200 rounded-xl px-6 py-10 text-center">
                        <p className="text-3xl mb-2">🧾</p>
                        <p className="text-gray-600 font-semibold">No orders in this period</p>
                    </div>
                ) : (
                    <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
                        <div className="px-5 py-3 border-b border-gray-100 flex items-center justify-between gap-3">
                            <p className="text-sm font-semibold text-gray-700">{orders.length} order{orders.length !== 1 ? "s" : ""}</p>
                            <div className="flex items-center gap-4">
                                <p className="text-sm font-bold text-gray-900">
                                    Total: ${orders.reduce((s, o) => s + o.total_price, 0).toFixed(2)}
                                </p>
                                <button
                                    onClick={handleSaveAll}
                                    disabled={savingAll}
                                    className="bg-gray-800 hover:bg-gray-900 text-white text-xs font-semibold px-4 py-2 rounded-lg disabled:opacity-50 transition-colors"
                                >
                                    {savingAll ? "Generating…" : "Save All as PDF"}
                                </button>
                            </div>
                        </div>
                        <table className="w-full text-sm">
                            <thead>
                            <tr className="bg-gray-50 text-gray-400 text-xs uppercase tracking-wide border-b border-gray-100">
                                <th className="text-left px-5 py-3 font-semibold">Order</th>
                                <th className="text-left px-4 py-3 font-semibold">Customer</th>
                                <th className="text-left px-4 py-3 font-semibold">Date</th>
                                <th className="text-left px-4 py-3 font-semibold">Status</th>
                                <th className="text-right px-4 py-3 font-semibold">Total</th>
                                <th className="px-5 py-3"></th>
                            </tr>
                            </thead>
                            <tbody>
                            {orders.map(order => <OrderRow key={order.delivery_id} order={order} />)}
                            </tbody>
                        </table>
                    </div>
                )
            )}

            {orders === null && !loading && !error && (
                <div className="bg-white border border-gray-200 rounded-xl px-6 py-14 text-center">
                    <p className="text-3xl mb-2">🧾</p>
                    <p className="text-gray-600 font-semibold">Select a date range and click Search</p>
                </div>
            )}
        </div>
    );
}