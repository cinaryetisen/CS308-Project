import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';

export default function ProductDetail() {
    const { id }   = useParams();
    const navigate = useNavigate();
    const API_URL  = import.meta.env.VITE_API_URL;

    const [product, setProduct]   = useState(null);
    const [loading, setLoading]   = useState(true);
    const [error, setError]       = useState("");
    const [quantity, setQuantity] = useState(1);
    const [cartMsg, setCartMsg]   = useState(null);

    // Auth
    const [isLoggedIn, setIsLoggedIn] = useState(false);

    // Reviews (comment only — goes to PM approval)
    const [reviews, setReviews]                 = useState([]);
    const [reviewsLoading, setReviewsLoading]   = useState(true);
    const [reviewError, setReviewError]         = useState("");
    const [reviewSubmitMsg, setReviewSubmitMsg] = useState(null);
    const [submittingReview, setSubmittingReview] = useState(false);
    const [newComment, setNewComment]           = useState("");

    // Ratings map: user_id -> rating number (for showing alongside comments)
    const [ratingsMap, setRatingsMap] = useState({});

    // Ratings (stars — instant, no approval)
    const [myRating, setMyRating]               = useState(0);
    const [hoveredStar, setHoveredStar]         = useState(0);
    const [ratingMsg, setRatingMsg]             = useState(null);
    const [submittingRating, setSubmittingRating] = useState(false);

    useEffect(() => {
        const token = localStorage.getItem("token");
        if (token) {
            setIsLoggedIn(true);
        }
    }, []);

    // Fetch product
    useEffect(() => {
        const fetchProduct = async () => {
            try {
                const res  = await fetch(`${API_URL}/api/products/${id}`);
                const text = await res.text();
                let data;
                try { data = JSON.parse(text); }
                catch { setError("Unexpected response from server."); return; }
                if (!res.ok) { setError(data.error || "Product not found."); return; }
                setProduct(data);
            } catch (err) {
                console.error(err);
                setError("Server error. Please try again later.");
            } finally {
                setLoading(false);
            }
        };
        fetchProduct();
    }, [id]);

    // Fetch approved comments
    useEffect(() => {
        const fetchReviews = async () => {
            try {
                const res = await fetch(`${API_URL}/api/products/${id}/reviews`);
                if (!res.ok) { setReviews([]); return; }
                const data = await res.json();
                setReviews(data || []);
            } catch (err) {
                console.error(err);
                setReviews([]);
            } finally {
                setReviewsLoading(false);
            }
        };
        fetchReviews();
    }, [id]);

    // Fetch all ratings for this product to show alongside comments
    useEffect(() => {
        const fetchRatings = async () => {
            try {
                const res = await fetch(`${API_URL}/api/products/${id}/ratings`);
                if (!res.ok) return;
                const data = await res.json();
                // Build a map of user_id -> rating for quick lookup
                const map = {};
                (data || []).forEach((r) => { map[r.user_id] = r.rating; });
                setRatingsMap(map);
            } catch (err) {
                console.error(err);
            }
        };
        fetchRatings();
    }, [id]);

    // Fetch user's existing rating to pre-fill stars
    useEffect(() => {
        if (!isLoggedIn) return;
        const fetchMyRating = async () => {
            try {
                const token = localStorage.getItem("token");
                const res = await fetch(`${API_URL}/api/me/ratings/${id}`, {
                    headers: { "Authorization": `Bearer ${token}` },
                });
                if (res.ok) {
                    const data = await res.json();
                    setMyRating(data.rating || 0);
                }
            } catch (err) {
                console.error(err);
            }
        };
        fetchMyRating();
    }, [id, isLoggedIn]);

    // Submit rating — instant, no approval
    const handleSubmitRating = async (star) => {
        if (!isLoggedIn || submittingRating) return;
        setMyRating(star);
        setSubmittingRating(true);
        setRatingMsg(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_URL}/api/ratings`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${token}`,
                },
                body: JSON.stringify({ product_id: id, rating: star }),
            });
            if (!res.ok) {
                let errMsg = "Failed to submit rating.";
                try { const data = await res.json(); errMsg = data.error || errMsg; } catch {}
                setRatingMsg({ type: "error", text: errMsg });
                return;
            }
            setRatingMsg({ type: "success", text: "Rating saved!" });
            setTimeout(() => setRatingMsg(null), 3000);
            // Refresh product to show updated average
            const productRes = await fetch(`${API_URL}/api/products/${id}`);
            if (productRes.ok) setProduct(await productRes.json());
        } catch (err) {
            console.error(err);
            setRatingMsg({ type: "error", text: "Server error. Please try again." });
        } finally {
            setSubmittingRating(false);
        }
    };

    // Submit comment — goes to PM for approval
    const handleSubmitReview = async () => {
        if (!newComment.trim()) { setReviewError("Please write a comment."); return; }
        setReviewError("");
        setSubmittingReview(true);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_URL}/api/reviews`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${token}`,
                },
                body: JSON.stringify({ product_id: id, comment: newComment }),
            });
            if (!res.ok) {
                let errMsg = "Failed to submit comment.";
                try { const data = await res.json(); errMsg = data.error || errMsg; } catch {}
                setReviewError(errMsg);
                return;
            }
            setNewComment("");
            setReviewSubmitMsg("Comment submitted! It will appear after moderation.");
            setTimeout(() => setReviewSubmitMsg(null), 4000);
        } catch (err) {
            console.error(err);
            setReviewError("Server error. Please try again.");
        } finally {
            setSubmittingReview(false);
        }
    };

    // Add to Cart
    const handleAddToCart = async () => {
        if (!product || product.quantity === 0) return;

        if (isLoggedIn) {
            try {
                const token = localStorage.getItem("token");
                const response = await fetch(`${API_URL}/api/cart/item`, {
                    method: 'PATCH',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token}`
                    },
                    body: JSON.stringify({ product_id: product.id, quantity: quantity })
                });
                if (response.ok) {
                    setCartMsg("added");
                    // Note: If you want the brown header to update its cart count, 
                    // you will need to trigger a global state update or context here.
                } else {
                    setCartMsg("maxed");
                }
                setTimeout(() => setCartMsg(null), 1500);
            } catch (error) {
                console.error("Network error adding to cart:", error);
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
            // Note: Dispatch an event here if you need the brown header to update instantly for guests
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
            <div className="flex justify-center items-center min-h-screen">
                <p className="text-gray-500">Loading product...</p>
            </div>
        );
    }

    if (error || !product) {
        return (
            <div className="flex flex-col items-center justify-center min-h-screen gap-4">
                <p className="text-red-500 text-lg">{error || "Product not found."}</p>
                <button onClick={() => navigate("/")} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
                    Back to Products
                </button>
            </div>
        );
    }

    const outOfStock      = product.quantity === 0;
    const discountedPrice = product.discount > 0 ? product.price * (1 - product.discount / 100) : null;
    const displayStar     = hoveredStar || myRating;

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">

            {/* Breadcrumb */}
            <div className="max-w-7xl mx-auto w-full px-4 sm:px-6 pt-5 text-sm text-gray-400 flex gap-2">
                <Link to="/" className="hover:text-blue-600">Home</Link>
                <span>/</span>
                <span className="text-gray-700 font-medium truncate">{product.name}</span>
            </div>

            {/* Product Detail */}
            <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 py-8 flex flex-col gap-8">
                <div className="bg-white rounded-2xl shadow-md overflow-hidden">
                    <div className="flex flex-col md:flex-row">

                        {/* LEFT: Image */}
                        <div className="md:w-1/2 bg-gray-100 flex items-center justify-center p-8 min-h-72 md:min-h-96">
                            {product.image_url ? (
                                <img src={product.image_url} alt={product.name} className="max-h-80 w-full object-contain" />
                            ) : (
                                <div className="text-gray-300 text-8xl">📦</div>
                            )}
                        </div>

                        {/* RIGHT: Info */}
                        <div className="md:w-1/2 p-6 sm:p-8 flex flex-col gap-4">

                            {product.category && (
                                <span className="text-xs text-blue-600 font-semibold uppercase tracking-wide">
                                    {product.category}
                                </span>
                            )}

                            <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 leading-snug">
                                {product.name}
                            </h1>

                            <div className="flex items-center gap-2">
                                {renderStars(product.rating, "text-lg")}
                                <span className="text-gray-500 text-sm">
                                    {product.rating?.toFixed(1)} ({product.review_count} ratings)
                                </span>
                            </div>

                            <div>
                                {discountedPrice ? (
                                    <div className="flex items-center gap-3">
                                        <p className="text-3xl font-bold text-blue-600">${discountedPrice.toFixed(2)}</p>
                                        <p className="text-lg text-gray-400 line-through">${Number(product.price).toFixed(2)}</p>
                                        <span className="text-sm text-white bg-red-500 px-2 py-0.5 rounded-full">-{product.discount}%</span>
                                    </div>
                                ) : (
                                    <p className="text-3xl font-bold text-blue-600">${Number(product.price).toFixed(2)}</p>
                                )}
                            </div>

                            <p className={`text-sm font-medium ${outOfStock ? "text-red-500" : "text-green-600"}`}>
                                {outOfStock ? "Out of Stock" : `${product.quantity} left in stock`}
                            </p>

                            {product.description && (
                                <p className="text-gray-600 text-sm leading-relaxed border-t pt-4">
                                    {product.description}
                                </p>
                            )}

                            <div className="text-xs text-gray-400 flex flex-col gap-1 border-t pt-3">
                                {product.model         && <span>Model: {product.model}</span>}
                                {product.serial_number && <span>Serial No: {product.serial_number}</span>}
                                {product.warranty      && <span>Warranty: {product.warranty}</span>}
                                {product.distributor   && <span>Distributor: {product.distributor}</span>}
                            </div>

                            {product.tags && product.tags.length > 0 && (
                                <div className="flex flex-wrap gap-2">
                                    {product.tags.map((tag) => (
                                        <span key={tag} className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded-full">
                                            {tag}
                                        </span>
                                    ))}
                                </div>
                            )}

                            {cartMsg === "added" && (
                                <div className="px-4 py-2 text-sm text-green-700 bg-green-100 border border-green-300 rounded-lg">
                                    ✓ Added to cart!
                                </div>
                            )}
                            {cartMsg === "maxed" && (
                                <div className="px-4 py-2 text-sm text-red-700 bg-red-100 border border-red-300 rounded-lg">
                                    Maximum stock reached.
                                </div>
                            )}

                            {!outOfStock && (
                                <div className="flex items-center gap-3">
                                    <span className="text-sm font-medium text-gray-700">Quantity:</span>
                                    <div className="flex items-center border border-gray-300 rounded-lg overflow-hidden">
                                        <button onClick={() => setQuantity((q) => Math.max(1, q - 1))} className="px-3 py-2 bg-gray-100 hover:bg-gray-200 font-bold text-gray-700 transition">−</button>
                                        <span className="px-4 py-2 text-sm font-semibold min-w-[2.5rem] text-center">{quantity}</span>
                                        <button onClick={() => setQuantity((q) => Math.min(product.quantity, q + 1))} className="px-3 py-2 bg-gray-100 hover:bg-gray-200 font-bold text-gray-700 transition">+</button>
                                    </div>
                                </div>
                            )}

                            <button
                                onClick={handleAddToCart}
                                disabled={outOfStock}
                                className={`w-full py-3 rounded-lg font-bold text-white transition ${
                                    outOfStock ? "bg-gray-200 text-gray-400 cursor-not-allowed" : "bg-blue-600 hover:bg-blue-700"
                                }`}
                            >
                                {outOfStock ? "Out of Stock" : "Add to Cart"}
                            </button>

                            <Link to="/" className="text-center text-sm text-gray-400 hover:text-blue-600 transition">
                                ← Back to Products
                            </Link>
                        </div>
                    </div>
                </div>

                {/* ── Ratings & Reviews Section ── */}
                <div className="bg-white rounded-2xl shadow-md p-6 sm:p-8 flex flex-col gap-6">
                    <h2 className="text-xl font-bold text-gray-900">
                        Ratings & Reviews
                        {reviews.length > 0 && (
                            <span className="ml-2 text-sm font-normal text-gray-400">({reviews.length} comments)</span>
                        )}
                    </h2>

                    {isLoggedIn ? (
                        <div className="border border-gray-200 rounded-xl p-5 flex flex-col gap-5 bg-gray-50">

                            {/* Star Rating — instant */}
                            <div className="flex flex-col gap-2">
                                <p className="text-sm font-semibold text-gray-700">
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
                                            className="text-2xl transition-transform hover:scale-110 focus:outline-none disabled:opacity-50"
                                        >
                                            <span className={displayStar >= star ? "text-yellow-400" : "text-gray-300"}>★</span>
                                        </button>
                                    ))}
                                    {myRating > 0 && (
                                        <span className="ml-2 text-sm text-gray-500">
                                            {["", "Poor", "Fair", "Good", "Very Good", "Excellent"][myRating]}
                                        </span>
                                    )}
                                    {submittingRating && <span className="ml-2 text-xs text-gray-400">Saving…</span>}
                                </div>
                                {ratingMsg && (
                                    <p className={`text-xs ${ratingMsg.type === "error" ? "text-red-600" : "text-green-600"}`}>
                                        {ratingMsg.text}
                                    </p>
                                )}
                            </div>

                            {/* Comment — goes to moderation */}
                            <div className="flex flex-col gap-2 border-t pt-4">
                                <p className="text-sm font-semibold text-gray-700">
                                    Leave a Comment
                                </p>
                                <textarea
                                    value={newComment}
                                    onChange={(e) => setNewComment(e.target.value)}
                                    placeholder="Share your experience with this product..."
                                    rows={3}
                                    className="w-full border border-gray-300 rounded-lg px-4 py-3 text-sm text-gray-700 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                />
                                {reviewError     && <p className="text-sm text-red-600">{reviewError}</p>}
                                {reviewSubmitMsg && <p className="text-sm text-green-600">✓ {reviewSubmitMsg}</p>}
                                <button
                                    onClick={handleSubmitReview}
                                    disabled={submittingReview}
                                    className="self-end px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {submittingReview ? "Submitting..." : "Submit Comment"}
                                </button>
                            </div>
                        </div>
                    ) : (
                        <div className="border border-dashed border-gray-300 rounded-xl p-5 text-center text-sm text-gray-500">
                            <Link to="/login" className="text-blue-600 hover:underline font-medium">Log in</Link>
                            {" "}to rate and review this product.
                        </div>
                    )}

                    {/* Approved comments list */}
                    {reviewsLoading ? (
                        <p className="text-sm text-gray-400">Loading reviews...</p>
                    ) : reviews.length === 0 ? (
                        <p className="text-sm text-gray-400">No approved reviews yet.</p>
                    ) : (
                        <div className="flex flex-col divide-y divide-gray-100">
                            {reviews.map((review) => (
                                <div key={review.id} className="py-5 flex flex-col gap-2">
                                    <div className="flex items-center gap-3">
                                        <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold text-sm flex-shrink-0">
                                            {review.user_name?.charAt(0).toUpperCase() || "?"}
                                        </div>
                                        <div>
                                            <p className="text-sm font-semibold text-gray-800">{review.user_name}</p>
                                            <p className="text-xs text-gray-400">{formatDate(review.created_at)}</p>
                                        </div>
                                    </div>
                                    {ratingsMap[review.user_id] && (
                                        <div className="flex items-center gap-1">
                                            {renderStars(ratingsMap[review.user_id], "text-sm")}
                                            <span className="text-xs text-gray-400">{["","Poor","Fair","Good","Very Good","Excellent"][ratingsMap[review.user_id]]}</span>
                                        </div>
                                    )}
                                    {review.comment && (
                                        <p className="text-sm text-gray-600 leading-relaxed pl-11">{review.comment}</p>
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