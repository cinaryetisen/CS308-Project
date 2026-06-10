import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link, useOutletContext } from 'react-router-dom';
import { apiRequest } from '../api/client';

export default function ProductDetail() {
    const { id }   = useParams();
    const navigate = useNavigate();

    const { refreshCartCount } = useOutletContext() || {};

    const [product, setProduct]   = useState(null);
    const [loading, setLoading]   = useState(true);
    const [error, setError]       = useState("");
    const [quantity, setQuantity] = useState(1);
    const [cartMsg, setCartMsg]   = useState(null);

    const [isLoggedIn, setIsLoggedIn] = useState(false);

    const [isWishlisted, setIsWishlisted]       = useState(false);
    const [wishlistLoading, setWishlistLoading] = useState(false);

    const [reviews, setReviews]                   = useState([]);
    const [reviewsLoading, setReviewsLoading]     = useState(true);
    const [reviewError, setReviewError]           = useState("");
    const [reviewSubmitMsg, setReviewSubmitMsg]   = useState(null);
    const [submittingReview, setSubmittingReview] = useState(false);
    const [newComment, setNewComment]             = useState("");

    const [ratingsMap, setRatingsMap] = useState({});

    const [myRating, setMyRating]                 = useState(0);
    const [hoveredStar, setHoveredStar]           = useState(0);
    const [ratingMsg, setRatingMsg]               = useState(null);
    const [submittingRating, setSubmittingRating] = useState(false);

    useEffect(() => {
        const token = localStorage.getItem("token");
        if (token) setIsLoggedIn(true);
    }, []);

    useEffect(() => {
        const fetchProduct = async () => {
            try {
                const data = await apiRequest(`/api/products/${id}`, {}, false);
                setProduct(data);
            } catch (err) {
                console.error(err);
                setError(err.message || "Product not found.");
            } finally {
                setLoading(false);
            }
        };
        fetchProduct();
    }, [id]);

    useEffect(() => {
        const fetchReviews = async () => {
            try {
                const data = await apiRequest(`/api/products/${id}/reviews`, {}, false);
                setReviews(data || []);
            } catch {
                setReviews([]);
            } finally {
                setReviewsLoading(false);
            }
        };
        fetchReviews();
    }, [id]);

    useEffect(() => {
        const fetchRatings = async () => {
            try {
                const data = await apiRequest(`/api/products/${id}/ratings`, {}, false);
                const map = {};
                (data || []).forEach((r) => { map[r.user_id] = r.rating; });
                setRatingsMap(map);
            } catch { /* non-critical */ }
        };
        fetchRatings();
    }, [id]);

    useEffect(() => {
        if (!isLoggedIn) return;
        const fetchMyRating = async () => {
            try {
                const data = await apiRequest(`/api/me/ratings/${id}`);
                setMyRating(data.rating || 0);
            } catch { /* no existing rating */ }
        };
        fetchMyRating();
    }, [id, isLoggedIn]);

    useEffect(() => {
        if (!isLoggedIn) return;
        const fetchWishlistStatus = async () => {
            try {
                const data = await apiRequest("/api/wishlist");
                setIsWishlisted(Array.isArray(data) && data.some((item) => String(item.id) === String(id)));
            } catch { /* non-critical */ }
        };
        fetchWishlistStatus();
    }, [id, isLoggedIn]);

    const handleSubmitRating = async (star) => {
        if (!isLoggedIn || submittingRating) return;
        setMyRating(star);
        setSubmittingRating(true);
        setRatingMsg(null);
        try {
            await apiRequest("/api/ratings", {
                method: "POST",
                body: JSON.stringify({ product_id: id, rating: star }),
            });
            setRatingMsg({ type: "success", text: "Rating saved!" });
            setTimeout(() => setRatingMsg(null), 3000);
            const updated = await apiRequest(`/api/products/${id}`, {}, false);
            setProduct(updated);
        } catch (err) {
            console.error(err);
            setRatingMsg({ type: "error", text: err.message || "Failed to submit rating." });
        } finally {
            setSubmittingRating(false);
        }
    };

    const handleSubmitReview = async () => {
        if (!newComment.trim()) { setReviewError("Please write a comment."); return; }
        setReviewError("");
        setSubmittingReview(true);
        try {
            await apiRequest("/api/reviews", {
                method: "POST",
                body: JSON.stringify({ product_id: id, comment: newComment }),
            });
            setNewComment("");
            setReviewSubmitMsg("Comment submitted! It will appear after moderation.");
            setTimeout(() => setReviewSubmitMsg(null), 4000);
        } catch (err) {
            console.error(err);
            setReviewError(err.message || "Failed to submit comment.");
        } finally {
            setSubmittingReview(false);
        }
    };

    const handleToggleWishlist = async () => {
        if (!isLoggedIn) { navigate("/login"); return; }
        if (wishlistLoading) return;
        setWishlistLoading(true);
        try {
            if (isWishlisted) {
                await apiRequest(`/api/wishlist/${id}`, { method: "DELETE" });
                setIsWishlisted(false);
            } else {
                await apiRequest("/api/wishlist", {
                    method: "POST",
                    body: JSON.stringify({ product_id: id }),
                });
                setIsWishlisted(true);
            }
        } catch (err) {
            console.error(err);
        } finally {
            setWishlistLoading(false);
        }
    };

    const handleAddToCart = async () => {
        if (!product || product.quantity === 0) return;
        if (isLoggedIn) {
            try {
                await apiRequest("/api/cart/item", {
                    method: "PATCH",
                    body: JSON.stringify({ product_id: product.id, quantity: quantity })
                });
                setCartMsg("added");
                if (refreshCartCount) refreshCartCount();
                window.dispatchEvent(new Event("cartUpdated"));
                setTimeout(() => setCartMsg(null), 1500);
            } catch (error) {
                console.error("Failed to add to cart:", error);
                setCartMsg("maxed");
                setTimeout(() => setCartMsg(null), 1500);
            }
        } else {
            const cart = JSON.parse(localStorage.getItem("cart") || "[]");
            const idx  = cart.findIndex((item) => item.id === product.id);
            if (idx >= 0) {
                if (cart[idx].quantity + quantity > product.quantity) {
                    setCartMsg("maxed");
                    setTimeout(() => setCartMsg(null), 1500);
                    return;
                }
                cart[idx].quantity = Math.min(cart[idx].quantity + quantity, product.quantity);
            } else {
                cart.push({
                    id:        product.id,
                    name:      product.name,
                    price:     product.price,
                    image_url: product.image_url,
                    stock:     product.quantity,
                    quantity:  Math.min(quantity, product.quantity),
                });
            }
            localStorage.setItem("cart", JSON.stringify(cart));
            setCartMsg("added");
            setTimeout(() => setCartMsg(null), 1500);
            if (refreshCartCount) refreshCartCount();
            window.dispatchEvent(new Event("cartUpdated"));
        }
    };

    const formatDate = (dateStr) => {
        const d = new Date(dateStr);
        return d.toLocaleDateString("en-US", { year: "numeric", month: "short", day: "numeric" });
    };

    const renderStars = (rating, size = "text-base") => (
        <span className={`${size} text-yellow-400`}>
            {"★".repeat(Math.floor(rating))}{"☆".repeat(5 - Math.floor(rating))}
        </span>
    );

    if (loading) {
        return (
            <div className="flex justify-center items-center min-h-screen bg-[#1a0f0a]">
                <p className="text-[#9a8c9b] tracking-widest animate-pulse">Summoning artifact…</p>
            </div>
        );
    }

    if (error || !product) {
        return (
            <div className="flex flex-col items-center justify-center min-h-screen bg-[#1a0f0a] gap-6">
                <p className="text-[#ffdad6] text-lg">{error || "Artifact not found."}</p>
                <button
                    onClick={() => navigate("/")}
                    className="px-6 py-2 rounded bg-gradient-to-r from-[#e7b4ff] to-[#8a47af] text-[#300049] font-semibold hover:brightness-110 transition"
                >
                    ← Return to the Vault
                </button>
            </div>
        );
    }

    const outOfStock      = product.quantity === 0;
    const discountedPrice = product.discount > 0 ? product.price * (1 - product.discount / 100) : null;
    const displayStar     = hoveredStar || myRating;

    return (
        <div className="min-h-screen bg-[#1a0f0a] flex flex-col">

            {/* Breadcrumb */}
            <div className="max-w-7xl mx-auto w-full px-4 sm:px-6 pt-5 text-sm text-[#9a8c9b] flex gap-2 items-center">
                <Link to="/" className="hover:text-[#e7b4ff] transition-colors">The Vault</Link>
                <span className="text-[#342720]">/</span>
                {product.category && (
                    <>
                        <Link
                            to={`/?category=${encodeURIComponent(product.category)}`}
                            className="hover:text-[#e7b4ff] transition-colors capitalize"
                        >
                            {product.category}
                        </Link>
                        <span className="text-[#342720]">/</span>
                    </>
                )}
                <span className="text-[#f5ded3] truncate font-medium">{product.name}</span>
            </div>

            <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 py-8 flex flex-col gap-8">

                {/* Product Card */}
                <div className="bg-[#251912] border border-[#342720] rounded-lg overflow-hidden shadow-[0_0_40px_rgba(138,71,175,0.08)]">
                    <div className="flex flex-col md:flex-row min-h-[500px]">

                        {/* LEFT — Image */}
                        <div className="md:w-1/2 relative overflow-hidden border-b md:border-b-0 md:border-r border-[#342720]">

                            <div className={`absolute top-4 right-4 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest shadow-sm z-10
                                ${outOfStock ? 'bg-[#93000a] text-[#ffdad6]' : 'bg-[#add461] text-[#131f00]'}`}>
                                {outOfStock ? "Out of Stock" : "In Stock"}
                            </div>

                            {discountedPrice && (
                                <div className="absolute top-4 left-4 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest bg-[#e7b4ff] text-[#300049] shadow-sm z-10">
                                    -{product.discount}%
                                </div>
                            )}

                            {product.image_url ? (
                                <img
                                    src={product.image_url}
                                    alt={product.name}
                                    className="w-full h-full object-cover min-h-[340px] md:min-h-[500px] transition-transform duration-500 hover:scale-105"
                                />
                            ) : (
                                <div className="w-full h-full min-h-[340px] md:min-h-[500px] flex items-center justify-center bg-[#1a0f0a]">
                                    <span className="text-[#342720] text-8xl">✦</span>
                                </div>
                            )}
                        </div>

                        {/* RIGHT — Info */}
                        <div className="md:w-1/2 p-6 sm:p-8 flex flex-col gap-4 overflow-y-auto">

                            {product.category && (
                                <span className="text-[10px] uppercase tracking-widest text-[#add461] font-semibold">
                                    {product.category}
                                </span>
                            )}

                            <h1 className="text-2xl sm:text-3xl font-serif text-[#f5ded3] leading-snug">
                                {product.name}
                            </h1>

                            <div className="flex items-center gap-2">
                                {renderStars(product.rating, "text-lg")}
                                <span className="text-[#9a8c9b] text-sm">
                                    {Number(product.rating).toFixed(1)} ({product.review_count} ratings)
                                </span>
                            </div>

                            <div>
                                {discountedPrice ? (
                                    <div className="flex items-center gap-3 flex-wrap">
                                        <span className="text-3xl font-bold text-[#e7b4ff]">${discountedPrice.toFixed(2)}</span>
                                        <span className="text-lg text-[#9a8c9b] line-through">${Number(product.price).toFixed(2)}</span>
                                        <span className="text-xs text-[#300049] bg-[#e7b4ff] px-2 py-0.5 rounded-full font-bold">
                                            -{product.discount}%
                                        </span>
                                    </div>
                                ) : (
                                    <span className="text-3xl font-bold text-[#e7b4ff]">${Number(product.price).toFixed(2)}</span>
                                )}
                            </div>

                            <p className={`text-sm font-medium ${outOfStock ? "text-[#ffdad6]" : "text-[#add461]"}`}>
                                {outOfStock ? "This artifact has left the vault." : `${product.quantity} remaining in the vault`}
                            </p>

                            {product.description && (
                                <p className="text-[#9a8c9b] text-sm leading-relaxed border-t border-[#342720] pt-4">
                                    {product.description}
                                </p>
                            )}

                            <div className="text-xs text-[#9a8c9b]/70 flex flex-col gap-1 border-t border-[#342720] pt-3">
                                {product.model         && <span>Model: <span className="text-[#9a8c9b]">{product.model}</span></span>}
                                {product.serial_number && <span>Serial No: <span className="text-[#9a8c9b]">{product.serial_number}</span></span>}
                                {product.warranty      && <span>Warranty: <span className="text-[#9a8c9b]">{product.warranty}</span></span>}
                                {product.distributor   && <span>Distributor: <span className="text-[#9a8c9b]">{product.distributor}</span></span>}
                            </div>

                            {product.tags && product.tags.length > 0 && (
                                <div className="flex flex-wrap gap-2">
                                    {product.tags.map((tag) => (
                                        <span
                                            key={tag}
                                            className="text-xs bg-[#1a0f0a] border border-[#342720] text-[#9a8c9b] px-3 py-1 rounded-full hover:border-[#8a47af] hover:text-[#e7b4ff] transition-colors"
                                        >
                                            {tag}
                                        </span>
                                    ))}
                                </div>
                            )}

                            {cartMsg === "added" && (
                                <div className="px-4 py-2 text-sm text-[#131f00] bg-[#add461] border border-[#add461]/50 rounded-lg font-medium">
                                    ✓ Artifact added to your collection!
                                </div>
                            )}
                            {cartMsg === "maxed" && (
                                <div className="px-4 py-2 text-sm text-[#ffdad6] bg-[#93000a]/40 border border-[#93000a] rounded-lg">
                                    Maximum stock reached.
                                </div>
                            )}

                            {!outOfStock && (
                                <div className="flex items-center gap-3">
                                    <span className="text-sm font-medium text-[#9a8c9b]">Quantity:</span>
                                    <div className="flex items-center border border-[#342720] rounded overflow-hidden">
                                        <button
                                            onClick={() => setQuantity((q) => Math.max(1, q - 1))}
                                            className="px-3 py-2 bg-[#1a0f0a] hover:bg-[#342720] text-[#e7b4ff] font-bold transition"
                                        >
                                            −
                                        </button>
                                        <span className="px-4 py-2 text-sm font-semibold min-w-[2.5rem] text-center text-[#f5ded3] bg-[#251912]">
                                            {quantity}
                                        </span>
                                        <button
                                            onClick={() => setQuantity((q) => Math.min(product.quantity, q + 1))}
                                            className="px-3 py-2 bg-[#1a0f0a] hover:bg-[#342720] text-[#e7b4ff] font-bold transition"
                                        >
                                            +
                                        </button>
                                    </div>
                                </div>
                            )}

                            {/* Wishlist button */}
                            <button
                                onClick={handleToggleWishlist}
                                disabled={wishlistLoading}
                                className={`w-full py-2.5 rounded font-semibold border transition-all duration-150 flex items-center justify-center gap-2 disabled:opacity-50 ${
                                    isWishlisted
                                        ? "border-[#8a47af] bg-[#8a47af]/10 text-[#e7b4ff] hover:bg-[#93000a]/10 hover:border-[#93000a] hover:text-[#ffdad6]"
                                        : "border-[#342720] text-[#9a8c9b] hover:border-[#8a47af] hover:text-[#e7b4ff]"
                                }`}
                            >
                                <span className="text-lg leading-none">
                                    {isWishlisted ? "♥" : "♡"}
                                </span>
                                {wishlistLoading
                                    ? "Saving…"
                                    : isWishlisted
                                        ? "Saved to Wishlist"
                                        : "Add to Wishlist"}
                            </button>

                            {/* Add to Cart */}
                            <button
                                onClick={handleAddToCart}
                                disabled={outOfStock}
                                className={`w-full py-3 rounded font-semibold active:scale-95 transition-all duration-150 shadow-md ${
                                    outOfStock
                                        ? "bg-[#342720] text-[#9a8c9b] cursor-not-allowed"
                                        : "bg-gradient-to-r from-[#e7b4ff] to-[#8a47af] text-[#300049] hover:brightness-110"
                                }`}
                            >
                                {outOfStock ? "Unavailable" : "Add to Cart"}
                            </button>

                            <Link
                                to="/"
                                className="text-center text-sm text-[#9a8c9b] hover:text-[#e7b4ff] transition-colors"
                            >
                                ← Return to the Vault
                            </Link>
                        </div>
                    </div>
                </div>

                {/* Ratings & Reviews */}
                <div className="bg-[#251912] border border-[#342720] rounded-lg p-6 sm:p-8 flex flex-col gap-6">

                    <h2 className="text-xl font-serif text-[#f5ded3]">
                        Ratings &amp; Reviews
                        {reviews.length > 0 && (
                            <span className="ml-2 text-sm font-normal text-[#9a8c9b]">({reviews.length} comments)</span>
                        )}
                    </h2>

                    {isLoggedIn ? (
                        <div className="border border-[#342720] rounded-lg p-5 flex flex-col gap-5 bg-[#1a0f0a]">

                            <div className="flex flex-col gap-2">
                                <p className="text-sm font-semibold text-[#9a8c9b] uppercase tracking-widest">
                                    Your Rating
                                </p>
                                <div className="flex items-center gap-1">
                                    {[1, 2, 3, 4, 5].map((star) => (
                                        <button
                                            key={star}
                                            type="button"
                                            onMouseEnter={() => setHoveredStar(star)}
                                            onMouseLeave={() => setHoveredStar(0)}
                                            onClick={() => handleSubmitRating(star)}
                                            disabled={submittingRating}
                                            className="text-2xl transition-transform hover:scale-125 focus:outline-none disabled:opacity-50"
                                        >
                                            <span className={displayStar >= star ? "text-yellow-400" : "text-[#342720]"}>★</span>
                                        </button>
                                    ))}
                                    {myRating > 0 && (
                                        <span className="ml-2 text-sm text-[#9a8c9b]">
                                            {["", "Poor", "Fair", "Good", "Very Good", "Excellent"][myRating]}
                                        </span>
                                    )}
                                    {submittingRating && (
                                        <span className="ml-2 text-xs text-[#9a8c9b] animate-pulse">Saving…</span>
                                    )}
                                </div>
                                {ratingMsg && (
                                    <p className={`text-xs ${ratingMsg.type === "error" ? "text-[#ffdad6]" : "text-[#add461]"}`}>
                                        {ratingMsg.text}
                                    </p>
                                )}
                            </div>

                            <div className="flex flex-col gap-2 border-t border-[#342720] pt-4">
                                <p className="text-sm font-semibold text-[#9a8c9b] uppercase tracking-widest">
                                    Leave a Comment
                                </p>
                                <textarea
                                    value={newComment}
                                    onChange={(e) => setNewComment(e.target.value)}
                                    placeholder="Share your experience with this artifact..."
                                    rows={3}
                                    className="w-full bg-[#251912] border border-[#342720] rounded px-4 py-3 text-sm text-[#f5ded3] placeholder-[#9a8c9b] resize-none focus:outline-none focus:border-[#e7b4ff] transition-colors"
                                />
                                {reviewError && (
                                    <p className="text-sm text-[#ffdad6]">{reviewError}</p>
                                )}
                                {reviewSubmitMsg && (
                                    <p className="text-sm text-[#add461]">✓ {reviewSubmitMsg}</p>
                                )}
                                <button
                                    onClick={handleSubmitReview}
                                    disabled={submittingReview}
                                    className="self-end px-6 py-2 bg-gradient-to-r from-[#e7b4ff] to-[#8a47af] text-[#300049] text-sm font-semibold rounded transition hover:brightness-110 disabled:opacity-50 disabled:cursor-not-allowed active:scale-95"
                                >
                                    {submittingReview ? "Submitting…" : "Submit Comment"}
                                </button>
                            </div>
                        </div>
                    ) : (
                        <div className="border border-dashed border-[#342720] rounded-lg p-5 text-center text-sm text-[#9a8c9b]">
                            <Link to="/login" className="text-[#e7b4ff] hover:underline font-medium">Log in</Link>
                            {" "}to rate and review this artifact.
                        </div>
                    )}

                    {reviewsLoading ? (
                        <p className="text-sm text-[#9a8c9b] animate-pulse">Loading reviews…</p>
                    ) : reviews.length === 0 ? (
                        <p className="text-sm text-[#9a8c9b]">No approved reviews yet. Be the first to speak.</p>
                    ) : (
                        <div className="flex flex-col divide-y divide-[#342720]">
                            {reviews.map((review) => (
                                <div key={review.id} className="py-5 flex flex-col gap-2">
                                    <div className="flex items-center gap-3">
                                        <div className="w-8 h-8 rounded-full bg-[#8a47af]/20 border border-[#8a47af]/40 flex items-center justify-center text-[#e7b4ff] font-bold text-sm flex-shrink-0">
                                            {review.user_name?.charAt(0).toUpperCase() || "?"}
                                        </div>
                                        <div>
                                            <p className="text-sm font-semibold text-[#f5ded3]">{review.user_name}</p>
                                            <p className="text-xs text-[#9a8c9b]">{formatDate(review.created_at)}</p>
                                        </div>
                                    </div>
                                    {ratingsMap[review.user_id] && (
                                        <div className="flex items-center gap-1 pl-11">
                                            {renderStars(ratingsMap[review.user_id], "text-sm")}
                                            <span className="text-xs text-[#9a8c9b]">
                                                {["", "Poor", "Fair", "Good", "Very Good", "Excellent"][ratingsMap[review.user_id]]}
                                            </span>
                                        </div>
                                    )}
                                    {review.comment && (
                                        <p className="text-sm text-[#9a8c9b] leading-relaxed pl-11">{review.comment}</p>
                                    )}
                                </div>
                            ))}
                        </div>
                    )}
                </div>

            </main>
        </div>
    );
}