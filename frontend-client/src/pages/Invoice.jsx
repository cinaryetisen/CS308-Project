import { useEffect, useState } from "react";
import { useParams, useNavigate, useLocation, Link } from "react-router-dom";

const API_URL = import.meta.env.VITE_API_URL;

export default function Invoice() {
    const { orderId }  = useParams();
    const navigate     = useNavigate();
    const location     = useLocation();

    const [order, setOrder]           = useState(location.state?.order || null);
    const [productMap, setProductMap] = useState({});
    const [user, setUser]             = useState(null);
    const [loading, setLoading]       = useState(!location.state?.order);
    const [error, setError]           = useState(null);
    const [downloading, setDownloading] = useState(false);

    // Fetch user info
    useEffect(() => {
        const token = localStorage.getItem("token");
        if (!token) { navigate("/login"); return; }
        fetch(`${API_URL}/api/users/me`, {
            headers: { Authorization: `Bearer ${token}` }
        })
            .then((r) => r.ok ? r.json() : Promise.reject())
            .then(setUser)
            .catch(() => {});
    }, [navigate]);

    // If we didn't get the order via router state, fetch it from orders list
    useEffect(() => {
        if (order) return;
        const token = localStorage.getItem("token");
        fetch(`${API_URL}/api/orders/me`, {
            headers: { Authorization: `Bearer ${token}` }
        })
            .then((r) => r.ok ? r.json() : Promise.reject(new Error("Failed to load order")))
            .then((orders) => {
                const found = orders.find((o) => String(o.delivery_id) === String(orderId));
                if (!found) throw new Error("Order not found");
                setOrder(found);
            })
            .catch((err) => setError(err.message))
            .finally(() => setLoading(false));
    }, [orderId, order]);

    // Fetch product names for all items
    useEffect(() => {
        if (!order?.items?.length) return;
        const isValidId = (id) => typeof id === "string" && /^[a-f0-9]{24}$/i.test(id);
        const uniqueIds = [...new Set(order.items.map((i) => i.product_id).filter(isValidId))];
        if (!uniqueIds.length) return;

        Promise.all(
            uniqueIds.map(async (pid) => {
                try {
                    const r = await fetch(`${API_URL}/api/products/${pid}`);
                    if (!r.ok) return [pid, null];
                    const p = await r.json();
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
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="flex flex-col items-center gap-3">
                    <div className="w-8 h-8 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin" />
                    <p className="text-sm text-gray-500">Loading invoice...</p>
                </div>
            </div>
        );
    }

    if (error || !order) {
        return (
            <div className="min-h-screen bg-gray-50 flex items-center justify-center">
                <div className="text-center">
                    <p className="text-red-500 mb-4">{error || "Invoice not found."}</p>
                    <button onClick={() => navigate("/")} className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700">
                        Back to Home
                    </button>
                </div>
            </div>
        );
    }

    const items = order.items || [];
    const date  = new Date(order.created_at).toLocaleDateString("en-US", {
        year: "numeric", month: "long", day: "numeric",
    });

    return (
        <div className="min-h-screen bg-gray-100 py-10 px-4">
            <div className="max-w-2xl mx-auto">

                {/* Success banner */}
                <div className="bg-green-50 border border-green-200 rounded-xl px-5 py-4 mb-6 flex items-start gap-3">
                    <span className="text-green-600 text-xl mt-0.5">✓</span>
                    <div>
                        <p className="text-green-800 font-semibold text-sm">Order placed successfully!</p>
                        {user && (
                            <p className="text-green-700 text-xs mt-0.5">
                                A PDF copy of this invoice has been sent to <strong>{user.email}</strong>.
                            </p>
                        )}
                    </div>
                </div>

                {/* Invoice card */}
                <div className="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden">

                    {/* Invoice header */}
                    <div className="px-6 py-5 border-b border-gray-100 flex items-center justify-between">
                        <div>
                            <h1 className="text-xl font-bold text-gray-900">Invoice</h1>
                            <p className="text-xs text-gray-400 mt-0.5">Order #{order.delivery_id} · {date}</p>
                        </div>
                        <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-green-100 text-green-700">
                            Confirmed
                        </span>
                    </div>

                    {/* Billed to / Ship to */}
                    <div className="grid grid-cols-2 gap-4 px-6 py-5 border-b border-gray-100">
                        <div>
                            <p className="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-1">Billed to</p>
                            {user ? (
                                <>
                                    <p className="text-sm font-medium text-gray-800">{user.name || "—"}</p>
                                    <p className="text-xs text-gray-500">{user.email}</p>
                                    {user.tax_id && (
                                        <p className="text-xs text-gray-500">Tax ID: {user.tax_id}</p>
                                    )}
                                </>
                            ) : (
                                <p className="text-sm text-gray-400 italic">Loading...</p>
                            )}
                        </div>
                        <div>
                            <p className="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-1">Ship to</p>
                            <p className="text-sm text-gray-800 whitespace-pre-line">{order.delivery_address}</p>
                        </div>
                    </div>

                    {/* Items table */}
                    <table className="w-full text-sm">
                        <thead>
                        <tr className="bg-gray-50 text-gray-400 text-xs uppercase tracking-wide border-b border-gray-100">
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
                                <tr key={item.id || item.product_id} className="border-b border-gray-100 last:border-0">
                                    <td className="px-6 py-3">
                                        {product ? (
                                            <Link
                                                to={`/products/${item.product_id}`}
                                                className="text-blue-600 hover:underline font-medium"
                                            >
                                                {product.name}
                                            </Link>
                                        ) : (
                                            <span className="text-gray-400 italic text-xs">{item.product_id}</span>
                                        )}
                                    </td>
                                    <td className="px-4 py-3 text-gray-600 text-center">{item.quantity}</td>
                                    <td className="px-4 py-3 text-gray-600 text-right">${item.price.toFixed(2)}</td>
                                    <td className="px-6 py-3 text-gray-900 font-semibold text-right">
                                        ${(item.price * item.quantity).toFixed(2)}
                                    </td>
                                </tr>
                            );
                        })}
                        </tbody>
                        <tfoot>
                        <tr className="border-t-2 border-gray-200">
                            <td colSpan="3" className="px-6 py-4 text-right text-sm font-semibold text-gray-700">
                                Total
                            </td>
                            <td className="px-6 py-4 text-right text-base font-bold text-gray-900">
                                ${order.total_price.toFixed(2)}
                            </td>
                        </tr>
                        </tfoot>
                    </table>

                    {/* Actions */}
                    <div className="px-6 py-5 border-t border-gray-100 flex justify-end gap-3">
                        <button
                            onClick={() => navigate("/")}
                            className="px-5 py-2 border border-gray-300 rounded-lg text-sm text-gray-600 hover:bg-gray-50 transition-colors"
                        >
                            Continue Shopping
                        </button>
                        <button
                            onClick={handleDownloadPdf}
                            disabled={downloading}
                            className="px-5 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-300 text-white rounded-lg text-sm font-semibold transition-colors"
                        >
                            {downloading ? "Opening..." : "Download PDF"}
                        </button>
                    </div>

                </div>

            </div>
        </div>
    );
}