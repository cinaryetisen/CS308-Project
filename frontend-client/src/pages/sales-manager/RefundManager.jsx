import { useState, useEffect } from "react";

const API_BASE = import.meta.env.VITE_API_URL;

const STATUS_STYLES = {
    pending:  "bg-yellow-100 text-yellow-700",
    approved: "bg-green-100 text-green-700",
    rejected: "bg-red-100 text-red-700",
};

function RefundCard({ refund, onResolved }) {
    const [saving, setSaving]   = useState(false);
    const [error, setError]     = useState(null);

    const date = new Date(refund.created_at).toLocaleDateString("en-US", {
        year: "numeric", month: "short", day: "numeric",
    });

    const resolvedDate = refund.resolved_at
        ? new Date(refund.resolved_at).toLocaleDateString("en-US", {
            year: "numeric", month: "short", day: "numeric",
          })
        : null;

    async function handleAction(action) {
        setSaving(true);
        setError(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/admin/refunds/${refund.id}`, {
                method: "PATCH",
                headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
                body: JSON.stringify({ action }),
            });
            if (!res.ok) { const e = await res.json(); throw new Error(e.error || "Failed"); }
            onResolved();
        } catch (e) {
            setError(e.message);
        } finally {
            setSaving(false);
        }
    }

    return (
        <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 px-5 py-4">
                <div className="flex flex-col gap-0.5">
                    <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
                        Refund #{refund.id}
                    </span>
                    <span className="text-sm text-gray-500">
                        Order #{refund.order_id} · Customer #{refund.customer_id} · {date}
                    </span>
                </div>
                <div className="flex items-center gap-3">
                    <span className={`text-xs font-semibold px-2.5 py-1 rounded-full ${STATUS_STYLES[refund.status] || "bg-gray-100 text-gray-600"}`}>
                        {refund.status.charAt(0).toUpperCase() + refund.status.slice(1)}
                    </span>
                    <span className="text-base font-bold text-gray-900">
                        ${Number(refund.refund_amount).toFixed(2)}
                    </span>
                </div>
            </div>

            {/* Reason */}
            <div className="px-5 pb-4 border-t border-gray-100 pt-3">
                <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
                    Reason
                </span>
                <p className="text-sm text-gray-700 mt-1">{refund.reason}</p>
            </div>

            {/* Resolved info */}
            {refund.status !== "pending" && resolvedDate && (
                <div className="px-5 pb-3 text-xs text-gray-400">
                    Resolved on {resolvedDate}
                    {refund.resolver_id && ` by Manager #${refund.resolver_id}`}
                </div>
            )}

            {/* Action buttons — only for pending */}
            {refund.status === "pending" && (
                <div className="px-5 py-4 border-t border-gray-100 bg-gray-50 flex items-center justify-between gap-3">
                    <p className="text-xs text-gray-400">
                        Approving will restore stock and mark the order as returned. The customer will be notified by email.
                    </p>
                    <div className="flex gap-2 shrink-0">
                        <button
                            onClick={() => handleAction("rejected")}
                            disabled={saving}
                            className="text-xs font-semibold px-4 py-2 border border-red-200 text-red-600 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-50"
                        >
                            Reject
                        </button>
                        <button
                            onClick={() => handleAction("approved")}
                            disabled={saving}
                            className="text-xs font-semibold px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg transition-colors disabled:opacity-50"
                        >
                            {saving ? "Saving…" : "Approve"}
                        </button>
                    </div>
                </div>
            )}

            {error && (
                <div className="px-5 py-2.5 text-sm text-red-600 bg-red-50 border-t border-red-100">
                    {error}
                </div>
            )}
        </div>
    );
}

export default function RefundManager() {
    const [refunds, setRefunds]   = useState([]);
    const [loading, setLoading]   = useState(true);
    const [error, setError]       = useState(null);
    const [filter, setFilter]     = useState("pending");

    useEffect(() => { fetchRefunds(); }, [filter]);

    async function fetchRefunds() {
        setLoading(true);
        setError(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/admin/refunds?status=${filter}`, {
                headers: { Authorization: `Bearer ${token}` },
            });
            if (!res.ok) throw new Error("Failed to fetch refunds");
            const data = await res.json();
            data.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
            setRefunds(data);
        } catch (e) {
            setError(e.message);
        } finally {
            setLoading(false);
        }
    }

    const FILTERS = ["pending", "approved", "rejected"];

    if (loading) {
        return (
            <div className="flex items-center justify-center py-24">
                <div className="flex flex-col items-center gap-3">
                    <div className="w-8 h-8 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin" />
                    <p className="text-sm text-gray-500">Loading refunds…</p>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-3xl mx-auto">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900">Refund Requests</h1>
                    <p className="text-sm text-gray-400 mt-0.5">
                        {refunds.length} {filter} request{refunds.length !== 1 ? "s" : ""}
                    </p>
                </div>
            </div>

            {/* Filter tabs */}
            <div className="flex gap-2 mb-4">
                {FILTERS.map(f => (
                    <button
                        key={f}
                        onClick={() => setFilter(f)}
                        className={`text-xs font-semibold px-3 py-1.5 rounded-lg border transition-colors capitalize ${
                            filter === f
                                ? "bg-blue-600 text-white border-blue-600"
                                : "bg-white text-gray-600 border-gray-300 hover:border-blue-400 hover:text-blue-600"
                        }`}
                    >
                        {f}
                    </button>
                ))}
            </div>

            {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3 mb-4">
                    {error}
                </div>
            )}

            {refunds.length === 0 ? (
                <div className="bg-white border border-gray-200 rounded-xl px-6 py-14 text-center">
                    <p className="text-4xl mb-3">
                        {filter === "pending" ? "✅" : filter === "approved" ? "💚" : "❌"}
                    </p>
                    <p className="text-gray-700 font-semibold">No {filter} refund requests</p>
                    <p className="text-sm text-gray-400 mt-1">
                        {filter === "pending" ? "All caught up!" : `No refunds have been ${filter} yet.`}
                    </p>
                </div>
            ) : (
                <div className="flex flex-col gap-3">
                    {refunds.map(refund => (
                        <RefundCard
                            key={refund.id}
                            refund={refund}
                            onResolved={fetchRefunds}
                        />
                    ))}
                </div>
            )}
        </div>
    );
}
