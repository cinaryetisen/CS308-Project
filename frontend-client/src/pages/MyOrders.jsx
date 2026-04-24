import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";

const API_BASE = import.meta.env.VITE_API_URL;

const STATUS_STYLES = {
    processing:   "bg-yellow-100 text-yellow-700",
    "in-transit": "bg-blue-100 text-blue-700",
    delivered:    "bg-green-100 text-green-700",
    cancelled:    "bg-red-100 text-red-700",
    returned:     "bg-gray-100 text-gray-600",
};

const STATUS_STEPS = ["processing", "in-transit", "delivered"];

function StatusBadge({ status }) {
    const style = STATUS_STYLES[status] || "bg-gray-100 text-gray-600";
    const label = status.charAt(0).toUpperCase() + status.slice(1);
    return (
        <span className={`text-xs font-semibold px-2.5 py-1 rounded-full ${style}`}>
      {label}
    </span>
    );
}

function StatusTracker({ status }) {
    const currentIndex = STATUS_STEPS.indexOf(status);
    return (
        <div className="flex items-center gap-0 w-full my-3">
            {STATUS_STEPS.map((step, i) => {
                const isCompleted = i < currentIndex;
                const isActive = i === currentIndex;
                return (
                    <div key={step} className="flex items-center flex-1">
                        <div className="flex flex-col items-center flex-1">
                            <div className={`w-3 h-3 rounded-full border-2 transition-colors ${
                                isCompleted ? "bg-blue-600 border-blue-600"
                                    : isActive   ? "bg-white border-blue-600"
                                        :              "bg-white border-gray-300"
                            }`} />
                            <span className={`text-xs mt-1 font-medium capitalize ${
                                isActive ? "text-blue-600" : isCompleted ? "text-gray-500" : "text-gray-300"
                            }`}>
                {step}
              </span>
                        </div>
                        {i < STATUS_STEPS.length - 1 && (
                            <div className={`h-0.5 flex-1 mb-4 ${i < currentIndex ? "bg-blue-600" : "bg-gray-200"}`} />
                        )}
                    </div>
                );
            })}
        </div>
    );
}

function OrderCard({ order, isCurrent }) {
    const [expanded, setExpanded] = useState(false);
    const date = new Date(order.created_at).toLocaleDateString("en-US", {
        year: "numeric", month: "short", day: "numeric",
    });

    return (
        <div className={`bg-white border rounded-xl overflow-hidden ${isCurrent ? "border-blue-200" : "border-gray-200"}`}>

            {/* Order header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 px-5 py-4">
                <div className="flex flex-col gap-0.5">
          <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
            Order #{order.delivery_id}
          </span>
                    <span className="text-sm text-gray-500">{date}</span>
                </div>
                <div className="flex items-center gap-3 sm:gap-4">
                    <StatusBadge status={order.status} />
                    <span className="text-base font-bold text-gray-900">
            ${order.total_price.toFixed(2)}
          </span>
                    <button
                        className="text-sm text-blue-600 hover:underline shrink-0"
                        onClick={() => setExpanded((e) => !e)}
                    >
                        {expanded ? "Hide items" : "View items"}
                    </button>
                </div>
            </div>

            {/* Status tracker — current orders only, and only for the 3 main statuses */}
            {isCurrent && STATUS_STEPS.includes(order.status) && (
                <div className="px-5 pb-2 border-t border-blue-50 pt-3">
          <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
            Delivery progress
          </span>
                    <StatusTracker status={order.status} />
                </div>
            )}

            {/* Delivery address */}
            <div className="px-5 pb-3 border-t border-gray-100 pt-3">
        <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide">
          Delivery address
        </span>
                <p className="text-sm text-gray-700 mt-0.5">{order.delivery_address}</p>
            </div>

            {/* Expandable items */}
            {expanded && (
                <div className="border-t border-gray-100">
                    {order.items && order.items.length > 0 ? (
                        <table className="w-full text-sm">
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
                    )}
                </div>
            )}
        </div>
    );
}

function Section({ title, count, children, emptyIcon, emptyTitle, emptyText }) {
    return (
        <div className="mb-8">
            <div className="flex items-center justify-between mb-3">
                <h2 className="text-lg font-bold text-gray-900">{title}</h2>
                {count > 0 && (
                    <span className="text-sm text-gray-400">{count} order{count !== 1 ? "s" : ""}</span>
                )}
            </div>
            {count === 0 ? (
                <div className="bg-white border border-gray-200 rounded-xl px-6 py-10 text-center">
                    <p className="text-3xl mb-2">{emptyIcon}</p>
                    <p className="text-gray-700 font-semibold">{emptyTitle}</p>
                    <p className="text-sm text-gray-400 mt-1">{emptyText}</p>
                </div>
            ) : (
                <div className="flex flex-col gap-3">{children}</div>
            )}
        </div>
    );
}

export default function MyOrders() {
    const [orders, setOrders] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const navigate = useNavigate();

    useEffect(() => {
        async function fetchOrders() {
            try {
                const token = localStorage.getItem("token");
                const res = await fetch(`${API_BASE}/api/orders/me`, {
                    headers: { Authorization: `Bearer ${token}` },
                });
                if (!res.ok) throw new Error("Failed to fetch orders");
                const data = await res.json();
                data.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
                setOrders(data);
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        }
        fetchOrders();
    }, []);

    if (loading) {
        return (
            <div className="min-h-screen bg-gray-100 flex items-center justify-center">
                <div className="flex flex-col items-center gap-3">
                    <div className="w-8 h-8 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin" />
                    <p className="text-sm text-gray-500">Loading your orders…</p>
                </div>
            </div>
        );
    }

    const currentOrders = orders.filter((o) => !o.completed);
    const previousOrders = orders.filter((o) => o.completed);

    return (
        <div className="min-h-screen bg-gray-100 py-8 px-4">
            <div className="max-w-2xl mx-auto">

                {/* Back link */}
                <button
                    className="text-blue-600 text-sm mb-5 hover:underline flex items-center gap-1"
                    onClick={() => navigate("/")}
                >
                    ← Back to products
                </button>

                {/* Page title */}
                <h1 className="text-2xl font-bold text-gray-900 mb-6">My Orders</h1>

                {/* Error state */}
                {error && (
                    <div className="bg-red-50 border border-red-200 text-red-700 text-sm font-medium rounded-lg px-4 py-3 mb-4">
                        {error}
                    </div>
                )}

                {!error && (
                    <>
                        {/* Current orders */}
                        <Section
                            title="Current Orders"
                            count={currentOrders.length}
                            emptyIcon="🚚"
                            emptyTitle="No active orders"
                            emptyText="Orders you place will appear here while they're being processed or delivered."
                        >
                            {currentOrders.map((order) => (
                                <OrderCard key={order.delivery_id} order={order} isCurrent={true} />
                            ))}
                        </Section>

                        {/* Previous orders */}
                        <Section
                            title="Order History"
                            count={previousOrders.length}
                            emptyIcon="📋"
                            emptyTitle="No previous orders"
                            emptyText="Completed and returned orders will appear here."
                        >
                            {previousOrders.map((order) => (
                                <OrderCard key={order.delivery_id} order={order} isCurrent={false} />
                            ))}
                        </Section>
                    </>
                )}

            </div>
        </div>
    );
}