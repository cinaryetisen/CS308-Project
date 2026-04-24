import { useState, useEffect } from "react";

const API_BASE = import.meta.env.VITE_API_URL;

const STATUSES = ["processing", "in-transit", "delivered"];

const STATUS_STYLES = {
    processing:   "bg-yellow-100 text-yellow-700",
    "in-transit": "bg-blue-100 text-blue-700",
    delivered:    "bg-green-100 text-green-700",
    cancelled:    "bg-red-100 text-red-700",
    returned:     "bg-gray-100 text-gray-600",
};

function StatusBadge({ status }) {
    const style = STATUS_STYLES[status] || "bg-gray-100 text-gray-600";
    const label = status.charAt(0).toUpperCase() + status.slice(1);
    return (
        <span className={`text-xs font-semibold px-2.5 py-1 rounded-full ${style}`}>
      {label}
    </span>
    );
}

function DeliveryCard({ order, onStatusUpdate }) {
    const [selected, setSelected] = useState(order.status);
    const [saving, setSaving] = useState(false);
    const [feedback, setFeedback] = useState(null);
    const [expanded, setExpanded] = useState(false);

    const date = new Date(order.created_at).toLocaleDateString("en-US", {
        year: "numeric", month: "short", day: "numeric",
    });

    const hasChanged = selected !== order.status;

    async function handleUpdate() {
        setSaving(true);
        setFeedback(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/deliveries/${order.delivery_id}/status`, {
                method: "PATCH",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${token}`,
                },
                body: JSON.stringify({ status: selected }),
            });
            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || "Update failed");
            }
            const data = await res.json();
            onStatusUpdate(order.delivery_id, data.order);
            setFeedback({ type: "success", message: "Status updated." });
        } catch (err) {
            setFeedback({ type: "error", message: err.message });
            setSelected(order.status);
        } finally {
            setSaving(false);
        }
    }

    return (
        <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">

            {/* Card header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 px-5 py-4">
                <div className="flex flex-col gap-0.5">
          <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
            Order #{order.delivery_id}
          </span>
                    <span className="text-sm text-gray-500">
            Customer #{order.customer_id} · {date}
          </span>
                </div>
                <div className="flex items-center gap-3">
                    <StatusBadge status={order.status} />
                    <span className="text-base font-bold text-gray-900">
            ${order.total_price.toFixed(2)}
          </span>
                </div>
            </div>

            {/* Delivery address */}
            <div className="px-5 pb-3 border-t border-gray-100 pt-3">
        <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
          Delivery address
        </span>
                <p className="text-sm text-gray-700 mt-0.5">{order.delivery_address}</p>
            </div>

            {/* Status update controls */}
            <div className="px-5 py-4 border-t border-gray-100 bg-gray-50 flex flex-col sm:flex-row sm:items-center gap-3">
                <label className="text-xs font-semibold text-gray-400 uppercase tracking-wide shrink-0">
                    Update status
                </label>
                <div className="flex flex-wrap gap-2 flex-1">
                    {STATUSES.map((s) => (
                        <button
                            key={s}
                            onClick={() => { setSelected(s); setFeedback(null); }}
                            className={`text-xs font-semibold px-3 py-1.5 rounded-lg border transition-colors ${
                                selected === s
                                    ? "bg-blue-600 text-white border-blue-600"
                                    : "bg-white text-gray-600 border-gray-300 hover:border-blue-400 hover:text-blue-600"
                            }`}
                        >
                            {s.charAt(0).toUpperCase() + s.slice(1)}
                        </button>
                    ))}
                </div>
                <button
                    onClick={handleUpdate}
                    disabled={!hasChanged || saving}
                    className="shrink-0 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-white text-sm font-semibold px-4 py-2 rounded-lg transition-colors"
                >
                    {saving ? "Saving…" : "Confirm"}
                </button>
            </div>

            {/* Feedback */}
            {feedback && (
                <div className={`px-5 py-2.5 text-sm font-medium border-t ${
                    feedback.type === "success"
                        ? "bg-green-50 text-green-700 border-green-100"
                        : "bg-red-50 text-red-700 border-red-100"
                }`}>
                    {feedback.message}
                </div>
            )}

            {/* Expandable items */}
            <div className="border-t border-gray-100">
                <button
                    className="w-full text-left px-5 py-2.5 text-xs font-semibold text-blue-600 hover:bg-gray-50 transition-colors"
                    onClick={() => setExpanded((e) => !e)}
                >
                    {expanded ? "▲ Hide items" : "▼ View items"} ({order.items?.length ?? 0})
                </button>
                {expanded && (
                    order.items && order.items.length > 0 ? (
                        <table className="w-full text-sm border-t border-gray-100">
                            <thead>
                            <tr className="bg-gray-50 text-gray-400 text-xs uppercase tracking-wide">
                                <th className="text-left px-5 py-2.5 font-semibold">Product ID</th>
                                <th className="text-center px-5 py-2.5 font-semibold">Qty</th>
                                <th className="text-right px-5 py-2.5 font-semibold">Unit Price</th>
                                <th className="text-right px-5 py-2.5 font-semibold">Subtotal</th>
                            </tr>
                            </thead>
                            <tbody>
                            {order.items.map((item) => (
                                <tr key={item.id} className="border-t border-gray-100">
                                    <td className="px-5 py-3 text-gray-700 font-medium">{item.product_id}</td>
                                    <td className="px-5 py-3 text-gray-600 text-center">{item.quantity}</td>
                                    <td className="px-5 py-3 text-gray-600 text-right">${item.price.toFixed(2)}</td>
                                    <td className="px-5 py-3 text-gray-900 font-semibold text-right">
                                        ${(item.price * item.quantity).toFixed(2)}
                                    </td>
                                </tr>
                            ))}
                            </tbody>
                        </table>
                    ) : (
                        <p className="text-sm text-gray-400 px-5 py-4 italic">No item details available.</p>
                    )
                )}
            </div>

        </div>
    );
}

export default function DeliveryManager() {
    const [orders, setOrders] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [filter, setFilter] = useState("all");

    useEffect(() => {
        async function fetchDeliveries() {
            try {
                const token = localStorage.getItem("token");
                const res = await fetch(`${API_BASE}/api/deliveries`, {
                    headers: { Authorization: `Bearer ${token}` },
                });
                if (!res.ok) throw new Error("Failed to fetch deliveries");
                const data = await res.json();
                data.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
                setOrders(data);
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        }
        fetchDeliveries();
    }, []);

    function handleStatusUpdate(orderId, updatedOrder) {
        setOrders((prev) =>
            prev.map((o) => (o.delivery_id === orderId ? updatedOrder : o))
        );
    }

    const filtered = filter === "all" ? orders : orders.filter((o) => o.status === filter);

    if (loading) {
        return (
            <div className="flex items-center justify-center py-24">
                <div className="flex flex-col items-center gap-3">
                    <div className="w-8 h-8 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin" />
                    <p className="text-sm text-gray-500">Loading deliveries…</p>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-2xl mx-auto">

            {/* Page title */}
            <div className="flex items-center justify-between mb-4">
                <h1 className="text-2xl font-bold text-gray-900">Deliveries</h1>
                {orders.length > 0 && (
                    <span className="text-sm text-gray-400">{filtered.length} of {orders.length} orders</span>
                )}
            </div>

            {/* Error state */}
            {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 text-sm font-medium rounded-lg px-4 py-3 mb-4">
                    {error}
                </div>
            )}

            {/* Filter tabs */}
            {!error && (
                <div className="flex gap-2 flex-wrap mb-4">
                    {["all", ...STATUSES].map((s) => (
                        <button
                            key={s}
                            onClick={() => setFilter(s)}
                            className={`text-xs font-semibold px-3 py-1.5 rounded-lg border transition-colors ${
                                filter === s
                                    ? "bg-blue-600 text-white border-blue-600"
                                    : "bg-white text-gray-600 border-gray-300 hover:border-blue-400 hover:text-blue-600"
                            }`}
                        >
                            {s === "all" ? "All" : s.charAt(0).toUpperCase() + s.slice(1)}
                            {" "}
                            <span className="opacity-70">
                ({s === "all" ? orders.length : orders.filter((o) => o.status === s).length})
              </span>
                        </button>
                    ))}
                </div>
            )}

            {/* Empty state */}
            {!error && filtered.length === 0 && (
                <div className="bg-white border border-gray-200 rounded-xl px-6 py-14 text-center">
                    <p className="text-4xl mb-3">📭</p>
                    <p className="text-gray-700 font-semibold">No orders found</p>
                    <p className="text-sm text-gray-400 mt-1">
                        {filter === "all" ? "There are no orders yet." : `No orders with status "${filter}".`}
                    </p>
                </div>
            )}

            {/* Orders list */}
            <div className="flex flex-col gap-3">
                {filtered.map((order) => (
                    <DeliveryCard
                        key={order.delivery_id}
                        order={order}
                        onStatusUpdate={handleStatusUpdate}
                    />
                ))}
            </div>

        </div>
    );
}