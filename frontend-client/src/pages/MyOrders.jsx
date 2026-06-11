import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { apiRequest } from "../api/client";

const STATUS_STYLES = {
    processing:   "bg-[#3a2800]/60 text-[#e7c46a] border border-[#5a4200]/50",
    "in-transit": "bg-[#1a1f3a]/60 text-[#b4d4ff] border border-[#2a3560]/50",
    delivered:    "bg-[#add461]/10 text-[#add461] border border-[#add461]/30",
    cancelled:    "bg-[#93000a]/20 text-[#ffdad6] border border-[#93000a]/40",
    returned:     "bg-[var(--surface-alt)]/60 text-[var(--muted)] border border-[var(--border)]",
};

const REFUND_STATUS_STYLES = {
    pending:  "bg-[#3a2800]/60 text-[#e7c46a] border border-[#5a4200]/50",
    approved: "bg-[#add461]/10 text-[#add461] border border-[#add461]/30",
    rejected: "bg-[#93000a]/20 text-[#ffdad6] border border-[#93000a]/40",
};

const STATUS_STEPS = ["processing", "in-transit", "delivered"];

function StatusBadge({ status }) {
    const style = STATUS_STYLES[status] || "bg-[var(--surface-alt)]/60 text-[var(--muted)] border border-[var(--border)]";
    const label = status.charAt(0).toUpperCase() + status.slice(1);
    return (
        <span className={`text-[10px] font-bold px-2.5 py-1 rounded-full uppercase tracking-widest ${style}`}>
            {label}
        </span>
    );
}

function StatusTracker({ status }) {
    const currentIndex = STATUS_STEPS.indexOf(status);
    return (
        <div className="flex items-center w-full my-3">
            {STATUS_STEPS.map((step, i) => {
                const isCompleted = i < currentIndex;
                const isActive    = i === currentIndex;
                return (
                    <div key={step} className="flex items-center flex-1">
                        <div className="flex flex-col items-center flex-1">
                            <div className={`w-3 h-3 rounded-full border-2 transition-colors ${
                                isCompleted ? "bg-[#add461] border-[#add461]"
                                    : isActive   ? "bg-transparent border-[var(--accent)]"
                                        :              "bg-transparent border-[var(--border)]"
                            }`} />
                            <span className={`text-[10px] mt-1 font-semibold capitalize tracking-wide ${
                                isActive ? "text-[var(--accent)]" : isCompleted ? "text-[#add461]" : "text-[var(--border)]"
                            }`}>
                                {step}
                            </span>
                        </div>
                        {i < STATUS_STEPS.length - 1 && (
                            <div className={`h-px flex-1 mb-4 ${i < currentIndex ? "bg-[#add461]" : "bg-[var(--border)]"}`} />
                        )}
                    </div>
                );
            })}
        </div>
    );
}

function RefundForm({ orderId, item, productMap, existingRefunds, onRefundSubmitted }) {
    const [reason, setReason]         = useState("");
    const [submitting, setSubmitting] = useState(false);
    const [msg, setMsg]               = useState(null);

    const existingRefund = existingRefunds.find((r) => r.order_item_id === item.id);
    const product        = productMap[item.product_id];

    if (existingRefund) {
        const refundStyle = REFUND_STATUS_STYLES[existingRefund.status] || "bg-[var(--surface-alt)]/60 text-[var(--muted)] border border-[var(--border)]";
        return (
            <div className="flex items-center justify-between px-4 py-2 bg-[var(--bg)] rounded border border-[var(--border)] text-sm mt-2">
                <span className="text-[var(--muted)] text-xs">Refund requested</span>
                <span className={`text-[10px] font-bold px-2.5 py-1 rounded-full uppercase tracking-widest ${refundStyle}`}>
                    {existingRefund.status.charAt(0).toUpperCase() + existingRefund.status.slice(1)}
                </span>
            </div>
        );
    }

    async function handleSubmit() {
        if (!reason.trim()) { setMsg({ type: "error", text: "Please enter a reason." }); return; }
        setSubmitting(true);
        setMsg(null);
        try {
            await apiRequest(`/api/orders/${orderId}/refund`, {
                method: "POST",
                body: JSON.stringify({ order_item_id: item.id, reason }),
            });
            setMsg({ type: "success", text: "Refund request submitted!" });
            setReason("");
            onRefundSubmitted();
        } catch (err) {
            setMsg({ type: "error", text: err.message });
        } finally {
            setSubmitting(false);
        }
    }

    return (
        <div className="mt-2 flex flex-col gap-2">
            <textarea
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder={`Reason for returning ${product?.name || "this item"}...`}
                rows={2}
                className="w-full bg-[var(--bg)] border border-[var(--border)] rounded px-3 py-2 text-sm text-[var(--text)] placeholder-[var(--muted)] resize-none focus:outline-none focus:border-[var(--accent)] transition-colors"
            />
            {msg && (
                <p className={`text-xs ${msg.type === "error" ? "text-[#ffdad6]" : "text-[#add461]"}`}>
                    {msg.type === "success" ? "✓ " : ""}{msg.text}
                </p>
            )}
            <button
                onClick={handleSubmit}
                disabled={submitting}
                className="self-end text-xs font-semibold px-4 py-1.5 rounded border border-[#93000a]/60 text-[#ffdad6] bg-[#93000a]/20 hover:bg-[#93000a]/40 disabled:opacity-40 disabled:cursor-not-allowed transition-all active:scale-95"
            >
                {submitting ? "Submitting…" : "Request Refund"}
            </button>
        </div>
    );
}

function ItemsTable({ order, productMap, existingRefunds, onRefundSubmitted }) {
    const items       = order.items;
    const isDelivered = order.status === "delivered";

    if (!items || items.length === 0) {
        return <p className="text-sm text-[var(--muted)] px-5 py-4 italic">No item details available.</p>;
    }

    return (
        <div className="w-full text-sm">
            <div className="grid grid-cols-4 text-[var(--muted)] text-[10px] uppercase tracking-widest px-5 py-2.5 font-semibold border-b border-[var(--border)]">
                <span>Product</span>
                <span className="text-center">Qty</span>
                <span className="text-right">Unit Price</span>
                <span className="text-right">Subtotal</span>
            </div>

            {items.map((item) => {
                const product = productMap[item.product_id];
                return (
                    <div key={item.id} className="border-t border-[var(--border)] px-5 py-3 flex flex-col gap-2">
                        <div className="grid grid-cols-4 items-center">
                            <span className="text-[var(--text)] font-medium">
                                {product
                                    ? <Link to={`/products/${item.product_id}`} className="hover:text-[var(--accent)] transition-colors">{product.name}</Link>
                                    : <span className="text-[var(--muted)] italic text-xs">{item.product_id}</span>
                                }
                            </span>
                            <span className="text-[var(--muted)] text-center">{item.quantity}</span>
                            <span className="text-[var(--muted)] text-right">${item.price.toFixed(2)}</span>
                            <span className="text-[var(--accent)] font-semibold text-right">${(item.price * item.quantity).toFixed(2)}</span>
                        </div>

                        {isDelivered && (
                            <RefundForm
                                orderId={order.delivery_id}
                                item={item}
                                productMap={productMap}
                                existingRefunds={existingRefunds}
                                onRefundSubmitted={onRefundSubmitted}
                            />
                        )}
                    </div>
                );
            })}
        </div>
    );
}

function OrderCard({ order, isCurrent, productMap, existingRefunds, onRefundSubmitted, onOrderChanged }) {
    const [expanded, setExpanded]     = useState(false);
    const [cancelling, setCancelling] = useState(false);
    const [cancelErr, setCancelErr]   = useState(null);
    const date = new Date(order.created_at).toLocaleDateString("en-US", {
        year: "numeric", month: "short", day: "numeric",
    });

    // Only orders still in "processing" can be cancelled (in-transit/delivered
    // must go through the refund flow). Matches the backend rule.
    const isCancellable = order.status === "processing";

    async function handleCancel() {
        if (!window.confirm(`Cancel order #${order.delivery_id}? This restocks the items and cannot be undone.`)) return;
        setCancelling(true);
        setCancelErr(null);
        try {
            await apiRequest(`/api/orders/${order.delivery_id}/cancel`, { method: "POST" });
            onOrderChanged();
        } catch (err) {
            setCancelErr(err.message);
            setCancelling(false);
        }
    }

    return (
        <div className={`bg-[var(--surface)] border rounded-lg overflow-hidden shadow-[0_0_20px_rgba(138,71,175,0.06)] ${
            isCurrent ? "border-[var(--accent-dim)]/40" : "border-[var(--border)]"
        }`}>

            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 px-5 py-4">
                <div className="flex flex-col gap-0.5">
                    <span className="text-[10px] font-semibold text-[var(--muted)] uppercase tracking-widest">
                        Order #{order.delivery_id}
                    </span>
                    <span className="text-sm text-[var(--muted)]">{date}</span>
                </div>
                <div className="flex items-center gap-3 sm:gap-4 flex-wrap">
                    <StatusBadge status={order.status} />
                    <span className="text-base font-bold text-[var(--accent)]">
                        ${order.total_price.toFixed(2)}
                    </span>
                    <button
                        className="text-sm text-[var(--muted)] hover:text-[var(--accent)] transition-colors shrink-0 underline underline-offset-2"
                        onClick={() => setExpanded((e) => !e)}
                    >
                        {expanded ? "Hide items" : "View items"}
                    </button>
                </div>
            </div>

            {/* Status tracker */}
            {isCurrent && STATUS_STEPS.includes(order.status) && (
                <div className="px-5 pb-2 border-t border-[var(--border)] pt-3">
                    <span className="text-[10px] font-semibold text-[var(--muted)] uppercase tracking-widest">
                        Delivery progress
                    </span>
                    <StatusTracker status={order.status} />
                </div>
            )}

            {/* Delivery address */}
            <div className="px-5 pb-3 border-t border-[var(--border)] pt-3">
                <span className="text-[10px] font-semibold text-[var(--muted)] uppercase tracking-widest">
                    Delivery address
                </span>
                <p className="text-sm text-[var(--text)] mt-1">{order.delivery_address}</p>
            </div>

            {/* Cancel (processing orders only) */}
            {isCancellable && (
                <div className="px-5 pb-3 flex flex-col items-end gap-1">
                    {cancelErr && <p className="text-xs text-[#ffdad6] self-start">✕ {cancelErr}</p>}
                    <button
                        onClick={handleCancel}
                        disabled={cancelling}
                        className="text-xs font-semibold px-4 py-1.5 rounded border border-[#93000a]/60 text-[#ffdad6] bg-[#93000a]/20 hover:bg-[#93000a]/40 disabled:opacity-40 disabled:cursor-not-allowed transition-all active:scale-95"
                    >
                        {cancelling ? "Cancelling…" : "Cancel Order"}
                    </button>
                </div>
            )}

            {/* Refund hint */}
            {order.status === "delivered" && !expanded && (
                <div className="px-5 pb-3">
                    <p className="text-xs text-[var(--muted)]/60">
                        Need to return an item?{" "}
                        <button onClick={() => setExpanded(true)} className="text-[var(--accent)] hover:underline">
                            View items
                        </button>{" "}
                        to request a refund.
                    </p>
                </div>
            )}

            {/* Expandable items */}
            {expanded && (
                <div className="border-t border-[var(--border)]">
                    <ItemsTable
                        order={order}
                        productMap={productMap}
                        existingRefunds={existingRefunds}
                        onRefundSubmitted={onRefundSubmitted}
                    />
                </div>
            )}
        </div>
    );
}

function Section({ title, count, children, emptyIcon, emptyTitle, emptyText }) {
    return (
        <div className="mb-8">
            <div className="flex items-center justify-between mb-3">
                <h2 className="text-xl font-serif text-[var(--text)]">{title}</h2>
                {count > 0 && (
                    <span className="text-sm text-[var(--muted)]">{count} order{count !== 1 ? "s" : ""}</span>
                )}
            </div>
            {count === 0 ? (
                <div className="bg-[var(--surface)] border border-[var(--border)] rounded-lg px-6 py-10 text-center">
                    <p className="text-3xl mb-2">{emptyIcon}</p>
                    <p className="text-[var(--text)] font-medium">{emptyTitle}</p>
                    <p className="text-sm text-[var(--muted)] mt-1">{emptyText}</p>
                </div>
            ) : (
                <div className="flex flex-col gap-3">{children}</div>
            )}
        </div>
    );
}

export default function MyOrders() {
    const [orders, setOrders]         = useState([]);
    const [productMap, setProductMap] = useState({});
    const [refunds, setRefunds]       = useState([]);
    const [loading, setLoading]       = useState(true);
    const [error, setError]           = useState(null);

    async function fetchRefunds() {
        try {
            const data = await apiRequest("/api/orders/me/refunds");
            setRefunds(data || []);
        } catch { /* silently fail */ }
    }

    async function fetchOrders() {
        try {
            const data = await apiRequest("/api/orders/me");
            data.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
            setOrders(data);

            const isValidId = (id) => typeof id === "string" && /^[a-f0-9]{24}$/i.test(id);
            const uniqueIds = [
                ...new Set(
                    data.flatMap((order) => (order.items || []).map((item) => item.product_id))
                        .filter(isValidId)
                ),
            ];

            if (uniqueIds.length > 0) {
                const results = await Promise.all(
                    uniqueIds.map(async (pid) => {
                        try {
                            const p = await apiRequest(`/api/products/${pid}`, {}, false);
                            return [pid, p];
                        } catch {
                            return [pid, null];
                        }
                    })
                );
                setProductMap(Object.fromEntries(results));
            }

            await fetchRefunds();
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    }

    useEffect(() => {
        fetchOrders();
    }, []);

    if (loading) {
        return (
            <div className="min-h-screen bg-[var(--bg)] flex items-center justify-center">
                <p className="text-[var(--muted)] tracking-widest animate-pulse">Summoning your orders…</p>
            </div>
        );
    }

    const currentOrders  = orders.filter((o) => !o.completed);
    const previousOrders = orders.filter((o) => o.completed);

    return (
        <div className="min-h-screen bg-[var(--bg)] flex flex-col">
            <main className="flex-1 max-w-2xl mx-auto w-full px-4 sm:px-6 py-8 sm:py-10 flex flex-col gap-2">

                {/* Breadcrumb */}
                <div className="text-sm text-[var(--muted)] flex gap-2 items-center mb-4">
                    <Link to="/" className="hover:text-[var(--accent)] transition-colors">The Vault</Link>
                    <span className="text-[var(--border)]">/</span>
                    <span className="text-[var(--text)] font-medium">My Orders</span>
                </div>

                <h1 className="text-4xl font-serif text-[var(--text)] mb-4">My Orders</h1>

                {error && (
                    <div className="bg-[#93000a]/20 border border-[#93000a]/50 text-[#ffdad6] text-sm font-medium rounded-lg px-4 py-3 mb-4">
                        ✕ {error}
                    </div>
                )}

                {!error && (
                    <>
                        <Section
                            title="Current Orders"
                            count={currentOrders.length}
                            emptyIcon="✦"
                            emptyTitle="No active orders"
                            emptyText="Orders you place will appear here while they're being processed or delivered."
                        >
                            {currentOrders.map((order) => (
                                <OrderCard
                                    key={order.delivery_id}
                                    order={order}
                                    isCurrent={true}
                                    productMap={productMap}
                                    existingRefunds={refunds}
                                    onRefundSubmitted={fetchRefunds}
                                    onOrderChanged={fetchOrders}
                                />
                            ))}
                        </Section>

                        <Section
                            title="Order History"
                            count={previousOrders.length}
                            emptyIcon="✦"
                            emptyTitle="No previous orders"
                            emptyText="Completed and returned orders will appear here."
                        >
                            {previousOrders.map((order) => (
                                <OrderCard
                                    key={order.delivery_id}
                                    order={order}
                                    isCurrent={false}
                                    productMap={productMap}
                                    existingRefunds={refunds}
                                    onRefundSubmitted={fetchRefunds}
                                    onOrderChanged={fetchOrders}
                                />
                            ))}
                        </Section>
                    </>
                )}

                <Link
                    to="/"
                    className="text-center text-sm text-[var(--muted)] hover:text-[var(--accent)] transition-colors mt-2"
                >
                    ← Return to the Vault
                </Link>

            </main>
        </div>
    );
}
