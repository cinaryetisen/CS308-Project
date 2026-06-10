import { useEffect, useState } from "react";
import { useParams, useNavigate, useLocation, Link } from "react-router-dom";
import { apiRequest } from "../api/client";

const API_URL = import.meta.env.VITE_API_URL;

export default function Invoice() {
    const { orderId }  = useParams();
    const navigate     = useNavigate();
    const location     = useLocation();

    const [order, setOrder]             = useState(location.state?.order || null);
    const [productMap, setProductMap]   = useState({});
    const [user, setUser]               = useState(null);
    const [loading, setLoading]         = useState(!location.state?.order);
    const [error, setError]             = useState(null);
    const [downloading, setDownloading] = useState(false);

    useEffect(() => {
        const token = localStorage.getItem("token");
        if (!token) { navigate("/login"); return; }
        apiRequest("/api/users/me").then(setUser).catch(() => {});
    }, [navigate]);

    useEffect(() => {
        if (order) return;
        apiRequest("/api/orders/me")
            .then((orders) => {
                const found = orders.find((o) => String(o.delivery_id) === String(orderId));
                if (!found) throw new Error("Order not found");
                setOrder(found);
            })
            .catch((err) => setError(err.message))
            .finally(() => setLoading(false));
    }, [orderId, order]);

    useEffect(() => {
        if (!order?.items?.length) return;
        const isValidId = (id) => typeof id === "string" && /^[a-f0-9]{24}$/i.test(id);
        const uniqueIds = [...new Set(order.items.map((i) => i.product_id).filter(isValidId))];
        if (!uniqueIds.length) return;

        Promise.all(
            uniqueIds.map(async (pid) => {
                try {
                    const p = await apiRequest(`/api/products/${pid}`, {}, false);
                    return [pid, p];
                } catch { return [pid, null]; }
            })
        ).then((results) => setProductMap(Object.fromEntries(results)));
    }, [order]);

    const handleDownloadPdf = async () => {
        setDownloading(true);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_URL}/api/orders/${orderId}/invoice`, {
                headers: { Authorization: `Bearer ${token}` }
            });
            if (!res.ok) throw new Error("Failed to download invoice.");
            const blob = await res.blob();
            const url  = URL.createObjectURL(blob);
            window.open(url, "_blank", "noopener,noreferrer");
        } catch (err) {
            alert(err.message);
        } finally {
            setDownloading(false);
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-[var(--bg)] flex items-center justify-center">
                <p className="text-[var(--muted)] tracking-widest animate-pulse">Summoning invoice…</p>
            </div>
        );
    }

    if (error || !order) {
        return (
            <div className="min-h-screen bg-[var(--bg)] flex flex-col items-center justify-center gap-6">
                <p className="text-[#ffdad6] text-lg">{error || "Invoice not found."}</p>
                <button
                    onClick={() => navigate("/")}
                    className="px-6 py-2 rounded bg-gradient-to-r from-[var(--btn-from)] to-[var(--btn-to)] text-[var(--on-accent)] font-semibold hover:brightness-110 transition"
                >
                    ← Return to the Vault
                </button>
            </div>
        );
    }

    const items = order.items || [];
    const date  = new Date(order.created_at).toLocaleDateString("en-US", {
        year: "numeric", month: "long", day: "numeric",
    });

    const labelClass = "text-[10px] uppercase tracking-widest font-semibold text-[var(--muted)]";

    return (
        <div className="min-h-screen bg-[var(--bg)] flex flex-col">
            <main className="flex-1 max-w-2xl mx-auto w-full px-4 sm:px-6 py-8 sm:py-10 flex flex-col gap-5">

                {/* Breadcrumb */}
                <div className="text-sm text-[var(--muted)] flex gap-2 items-center">
                    <Link to="/" className="hover:text-[var(--accent)] transition-colors">The Vault</Link>
                    <span className="text-[var(--border)]">/</span>
                    <Link to="/orders" className="hover:text-[var(--accent)] transition-colors">My Orders</Link>
                    <span className="text-[var(--border)]">/</span>
                    <span className="text-[var(--text)] font-medium">Invoice</span>
                </div>

                {/* Success banner */}
                <div className="bg-[#add461]/10 border border-[#add461]/30 rounded-lg px-5 py-4 flex items-start gap-3">
                    <span className="text-[#add461] text-lg mt-0.5 leading-none">✓</span>
                    <div>
                        <p className="text-[#add461] font-semibold text-sm">Order placed successfully!</p>
                        {user && (
                            <p className="text-[#add461]/70 text-xs mt-0.5">
                                A PDF copy of this invoice has been sent to <strong>{user.email}</strong>.
                            </p>
                        )}
                    </div>
                </div>

                {/* Invoice card */}
                <div className="bg-[var(--surface)] border border-[var(--border)] rounded-lg overflow-hidden shadow-[0_0_40px_rgba(138,71,175,0.08)]">

                    {/* Invoice header */}
                    <div className="px-6 py-5 border-b border-[var(--border)] flex items-center justify-between">
                        <div>
                            <h1 className="text-2xl font-serif text-[var(--text)]">Invoice</h1>
                            <p className="text-xs text-[var(--muted)] mt-0.5">Order #{order.delivery_id} · {date}</p>
                        </div>
                        <span className="text-[10px] font-bold px-2.5 py-1 rounded-full uppercase tracking-widest bg-[#add461]/10 text-[#add461] border border-[#add461]/30">
                            Confirmed
                        </span>
                    </div>

                    {/* Billed to / Ship to */}
                    <div className="grid grid-cols-2 gap-4 px-6 py-5 border-b border-[var(--border)]">
                        <div>
                            <p className={`${labelClass} mb-1`}>Billed to</p>
                            {user ? (
                                <>
                                    <p className="text-sm font-medium text-[var(--text)]">{user.name || "—"}</p>
                                    <p className="text-xs text-[var(--muted)]">{user.email}</p>
                                    {user.tax_id && (
                                        <p className="text-xs text-[var(--muted)]">Tax ID: {user.tax_id}</p>
                                    )}
                                </>
                            ) : (
                                <p className="text-sm text-[var(--muted)] italic animate-pulse">Loading…</p>
                            )}
                        </div>
                        <div>
                            <p className={`${labelClass} mb-1`}>Ship to</p>
                            <p className="text-sm text-[var(--text)] whitespace-pre-line">{order.delivery_address}</p>
                        </div>
                    </div>

                    {/* Items table */}
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="text-[var(--muted)] text-[10px] uppercase tracking-widest border-b border-[var(--border)]">
                                <th className="text-left px-6 py-3 font-semibold">Product</th>
                                <th className="text-center px-4 py-3 font-semibold">Qty</th>
                                <th className="text-right px-4 py-3 font-semibold">Unit Price</th>
                                <th className="text-right px-6 py-3 font-semibold">Subtotal</th>
                            </tr>
                        </thead>
                        <tbody>
                            {items.map((item) => {
                                const product = productMap[item.product_id];
                                return (
                                    <tr key={item.id || item.product_id} className="border-b border-[var(--border)] last:border-0">
                                        <td className="px-6 py-3">
                                            {product ? (
                                                <Link
                                                    to={`/products/${item.product_id}`}
                                                    className="text-[var(--text)] font-medium hover:text-[var(--accent)] transition-colors"
                                                >
                                                    {product.name}
                                                </Link>
                                            ) : (
                                                <span className="text-[var(--muted)] italic text-xs">{item.product_id}</span>
                                            )}
                                        </td>
                                        <td className="px-4 py-3 text-[var(--muted)] text-center">{item.quantity}</td>
                                        <td className="px-4 py-3 text-[var(--muted)] text-right">${item.price.toFixed(2)}</td>
                                        <td className="px-6 py-3 text-[var(--accent)] font-semibold text-right">
                                            ${(item.price * item.quantity).toFixed(2)}
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                        <tfoot>
                            <tr className="border-t-2 border-[var(--border)]">
                                <td colSpan="3" className="px-6 py-4 text-right text-sm font-semibold text-[var(--muted)]">
                                    Total
                                </td>
                                <td className="px-6 py-4 text-right text-lg font-bold text-[var(--accent)]">
                                    ${order.total_price.toFixed(2)}
                                </td>
                            </tr>
                        </tfoot>
                    </table>

                    {/* Actions */}
                    <div className="px-6 py-5 border-t border-[var(--border)] flex justify-end gap-3">
                        <button
                            onClick={() => navigate("/")}
                            className="px-5 py-2 text-sm font-medium rounded border border-[var(--border)] text-[var(--muted)] hover:border-[var(--accent-dim)] hover:text-[var(--accent)] transition-all"
                        >
                            Continue Shopping
                        </button>
                        <button
                            onClick={handleDownloadPdf}
                            disabled={downloading}
                            className="px-5 py-2 text-sm font-semibold rounded bg-gradient-to-r from-[var(--btn-from)] to-[var(--btn-to)] text-[var(--on-accent)] hover:brightness-110 active:scale-95 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {downloading ? "Opening…" : "Download PDF"}
                        </button>
                    </div>

                </div>

                <Link
                    to="/orders"
                    className="text-center text-sm text-[var(--muted)] hover:text-[var(--accent)] transition-colors"
                >
                    ← Back to My Orders
                </Link>

            </main>
        </div>
    );
}
