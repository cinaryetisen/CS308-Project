import { useEffect, useState } from 'react';
import { Link, useNavigate, useOutletContext } from 'react-router-dom';

export default function Wishlist() {
    const API_URL  = import.meta.env.VITE_API_URL;
    const navigate = useNavigate();

    const { refreshCartCount, refreshWishlistCount } = useOutletContext() || {};

    const [wishlist, setWishlist]         = useState([]);
    const [loading, setLoading]           = useState(true);
    const [isLoggedIn, setIsLoggedIn]     = useState(false);
    const [cartFeedback, setCartFeedback] = useState({});
    const [removingId, setRemovingId]     = useState(null);

    useEffect(() => {
        const token = localStorage.getItem("token");
        if (token) {
            setIsLoggedIn(true);
        } else {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        if (!isLoggedIn) return;
        const fetchWishlist = async () => {
            setLoading(true);
            try {
                const token = localStorage.getItem("token");
                const res   = await fetch(`${API_URL}/api/wishlist`, {
                    headers: { "Authorization": `Bearer ${token}` },
                });
                if (!res.ok) { setWishlist([]); return; }
                const data = await res.json();
                setWishlist(Array.isArray(data) ? data : []);
            } catch (err) {
                console.error(err);
                setWishlist([]);
            } finally {
                setLoading(false);
            }
        };
        fetchWishlist();
    }, [isLoggedIn]);

    // Safely get the product id regardless of API response shape
    const getProductId = (item) => {
        return item.product_id ?? item.id ?? null;
    };

    const removeFromWishlist = async (item) => {
        const productId = getProductId(item);
        if (!productId) return;
        setRemovingId(productId);
        try {
            const token = localStorage.getItem("token");
            const res   = await fetch(`${API_URL}/api/wishlist/${productId}`, {
                method: "DELETE",
                headers: { "Authorization": `Bearer ${token}` },
            });
            if (res.ok) {
                setWishlist((prev) => prev.filter((i) => getProductId(i) !== productId));
                if (refreshWishlistCount) refreshWishlistCount();
            }
        } catch (err) {
            console.error(err);
        } finally {
            setRemovingId(null);
        }
    };

    const addToCart = async (e, item) => {
        e.stopPropagation();
        const productId = getProductId(item);
        if (!productId || item.quantity === 0) return;
        try {
            const token    = localStorage.getItem("token");
            const response = await fetch(`${API_URL}/api/cart/item`, {
                method:  "PATCH",
                headers: {
                    "Content-Type":  "application/json",
                    "Authorization": `Bearer ${token}`,
                },
                body: JSON.stringify({ product_id: productId, quantity: 1 }),
            });
            if (response.ok) {
                setCartFeedback((prev) => ({ ...prev, [productId]: "added" }));
                setTimeout(() => setCartFeedback((prev) => ({ ...prev, [productId]: null })), 1200);
                if (refreshCartCount) refreshCartCount();
            } else {
                setCartFeedback((prev) => ({ ...prev, [productId]: "maxed" }));
                setTimeout(() => setCartFeedback((prev) => ({ ...prev, [productId]: null })), 1500);
            }
        } catch (err) {
            console.error(err);
        }
    };

    // ── Not logged in ─────────────────────────────────────────────────────────
    if (!isLoggedIn) {
        return (
            <div className="min-h-screen bg-[var(--bg)] flex flex-col items-center justify-center gap-6 px-4">
                <div className="text-6xl">🔮</div>
                <h2 className="text-2xl font-serif text-[var(--text)]">Your Wishlist Awaits</h2>
                <p className="text-[var(--muted)] text-sm text-center max-w-xs">
                    Sign in to save artifacts to your wishlist and revisit them anytime.
                </p>
                <div className="flex gap-3">
                    <Link
                        to="/login"
                        className="px-6 py-2 rounded bg-gradient-to-r from-[var(--btn-from)] to-[var(--btn-to)] text-[var(--on-accent)] font-semibold hover:brightness-110 transition"
                    >
                        Log In
                    </Link>
                    <Link
                        to="/"
                        className="px-6 py-2 rounded border border-[var(--border)] text-[var(--muted)] hover:border-[var(--accent-dim)] hover:text-[var(--accent)] transition"
                    >
                        Browse the Vault
                    </Link>
                </div>
            </div>
        );
    }

    // ── Loading ───────────────────────────────────────────────────────────────
    if (loading) {
        return (
            <div className="min-h-screen bg-[var(--bg)] flex items-center justify-center">
                <p className="text-[var(--muted)] tracking-widest animate-pulse">Gathering your artifacts…</p>
            </div>
        );
    }

    // ── Empty ─────────────────────────────────────────────────────────────────
    if (wishlist.length === 0) {
        return (
            <div className="min-h-screen bg-[var(--bg)] flex flex-col">
                <div className="max-w-7xl mx-auto w-full px-4 sm:px-6 pt-5 text-sm text-[var(--muted)] flex gap-2 items-center">
                    <Link to="/" className="hover:text-[var(--accent)] transition-colors">The Vault</Link>
                    <span className="text-[var(--border)]">/</span>
                    <span className="text-[var(--text)] font-medium">Wishlist</span>
                </div>
                <div className="flex-1 flex flex-col items-center justify-center gap-6 px-4">
                    <div className="text-6xl opacity-30">✦</div>
                    <h2 className="text-2xl font-serif text-[var(--text)]">Your wishlist is empty</h2>
                    <p className="text-[var(--muted)] text-sm text-center max-w-xs">
                        Wander the vault and save the artifacts that call to you.
                    </p>
                    <Link
                        to="/"
                        className="px-6 py-2 rounded bg-gradient-to-r from-[var(--btn-from)] to-[var(--btn-to)] text-[var(--on-accent)] font-semibold hover:brightness-110 transition"
                    >
                        Browse the Vault
                    </Link>
                </div>
            </div>
        );
    }

    // ── Main ──────────────────────────────────────────────────────────────────
    return (
        <div className="min-h-screen bg-[var(--bg)] flex flex-col">

            {/* Breadcrumb */}
            <div className="max-w-7xl mx-auto w-full px-4 sm:px-6 pt-5 text-sm text-[var(--muted)] flex gap-2 items-center">
                <Link to="/" className="hover:text-[var(--accent)] transition-colors">The Vault</Link>
                <span className="text-[var(--border)]">/</span>
                <span className="text-[var(--text)] font-medium">Wishlist</span>
            </div>

            <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 py-8 sm:py-10">

                {/* Header */}
                <div className="flex items-center justify-between mb-8">
                    <h2 className="text-4xl font-serif text-[var(--text)]">Your Wishlist</h2>
                    <span className="text-sm text-[var(--muted)]">
                        {wishlist.length} {wishlist.length === 1 ? "artifact" : "artifacts"}
                    </span>
                </div>

                {/* Grid */}
                <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-8">
                    {wishlist.map((item) => {
                        const productId       = getProductId(item);
                        const outOfStock      = item.quantity === 0;
                        const feedback        = cartFeedback[productId];
                        const discountedPrice = item.discount > 0
                            ? item.price * (1 - item.discount / 100)
                            : null;
                        const isRemoving      = removingId === productId;

                        // rating may be missing or a string — guard both
                        const rating      = item.rating != null ? Number(item.rating) : null;
                        const reviewCount = item.review_count ?? null;

                        return (
                            <div
                                key={productId}
                                className="group relative overflow-hidden bg-[var(--surface)] border border-[var(--border)] hover:border-[var(--accent-dim)] hover:shadow-[0_0_20px_rgba(138,71,175,0.15)] rounded-lg transition-all duration-300 flex flex-col"
                            >
                                {/* Hover glow */}
                                <div className="absolute inset-0 bg-gradient-to-r from-[var(--accent)]/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none" />

                                {/* Remove X button */}
                                <button
                                    onClick={() => removeFromWishlist(item)}
                                    disabled={isRemoving}
                                    className="absolute top-3 right-3 z-20 w-7 h-7 flex items-center justify-center rounded-full bg-[var(--bg)]/80 border border-[var(--border)] text-[var(--muted)] hover:border-[#93000a] hover:text-[#ffdad6] transition-all disabled:opacity-40"
                                    title="Remove from wishlist"
                                >
                                    {isRemoving
                                        ? <span className="text-[10px] animate-pulse">…</span>
                                        : <span className="text-sm leading-none">✕</span>
                                    }
                                </button>

                                {/* Image */}
                                <div
                                    onClick={() => productId && navigate(`/products/${productId}`)}
                                    className="relative z-10 aspect-[4/5] overflow-hidden border-b border-[var(--border)]/50 cursor-pointer"
                                >
                                    <img
                                        src={item.image_url}
                                        alt={item.name}
                                        className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                                    />

                                    {/* Stock badge */}
                                    <div className={`absolute top-3 left-3 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest shadow-sm
                                        ${outOfStock ? 'bg-[#93000a] text-[#ffdad6]' : 'bg-[#add461] text-[#131f00]'}`}>
                                        {outOfStock ? "Out of Stock" : "In Stock"}
                                    </div>

                                    {/* Discount badge */}
                                    {discountedPrice && (
                                        <div className="absolute bottom-3 left-3 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest bg-[var(--btn-from)] text-[var(--on-accent)] shadow-sm">
                                            -{item.discount}%
                                        </div>
                                    )}
                                </div>

                                {/* Info */}
                                <div
                                    onClick={() => productId && navigate(`/products/${productId}`)}
                                    className="relative z-10 px-4 pt-3 pb-1 flex flex-col gap-1 cursor-pointer"
                                >
                                    <span className="text-[10px] uppercase tracking-widest text-[#add461]">
                                        {item.category || "Artifact"}
                                    </span>
                                    <h3 className="text-lg font-serif text-[var(--text)] line-clamp-2">
                                        {item.name}
                                    </h3>

                                    {/* Rating — only render if data exists */}
                                    {rating != null && !isNaN(rating) ? (
                                        <div className="text-sm text-[var(--muted)]">
                                            <span className="text-[#add461]">⭐ {rating.toFixed(1)}</span>
                                            {reviewCount != null && (
                                                <span> ({reviewCount} reviews)</span>
                                            )}
                                        </div>
                                    ) : null}

                                    <div className="mt-1 mb-3">
                                        {discountedPrice ? (
                                            <div className="flex items-center gap-2">
                                                <span className="text-[var(--accent)] font-bold">${discountedPrice.toFixed(2)}</span>
                                                <span className="text-sm text-[var(--muted)] line-through">${Number(item.price).toFixed(2)}</span>
                                            </div>
                                        ) : (
                                            <span className="text-[var(--accent)] font-bold">${Number(item.price).toFixed(2)}</span>
                                        )}
                                    </div>
                                </div>

                                {/* Actions */}
                                <div className="relative z-10 px-4 pb-4 flex flex-col gap-2 mt-auto">
                                    <button
                                        onClick={(e) => addToCart(e, item)}
                                        disabled={outOfStock}
                                        className={`w-full py-2 rounded font-semibold active:scale-95 transition-all duration-150 shadow-md ${
                                            outOfStock
                                                ? "bg-[var(--surface-alt)] text-[var(--muted)] cursor-not-allowed"
                                                : feedback === "added"
                                                    ? "bg-[#add461] text-[#131f00]"
                                                    : feedback === "maxed"
                                                        ? "bg-[#93000a] text-[#ffdad6]"
                                                        : "bg-gradient-to-r from-[var(--btn-from)] to-[var(--btn-to)] text-[var(--on-accent)] hover:brightness-110"
                                        }`}
                                    >
                                        {outOfStock
                                            ? "Unavailable"
                                            : feedback === "added"
                                                ? "✓ Added"
                                                : feedback === "maxed"
                                                    ? "Max Reached"
                                                    : "Add to Cart"}
                                    </button>

                                    <button
                                        onClick={() => removeFromWishlist(item)}
                                        disabled={isRemoving}
                                        className="w-full py-2 rounded text-sm text-[var(--muted)] border border-[var(--border)] hover:border-[#93000a] hover:text-[#ffdad6] transition-all disabled:opacity-40"
                                    >
                                        {isRemoving ? "Removing…" : "Remove from Wishlist"}
                                    </button>
                                </div>
                            </div>
                        );
                    })}
                </div>
            </main>
        </div>
    );
}
