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
    const [isLoggedIn, setIsLoggedIn]     = useState(false);
    const [userData, setUserData]         = useState(null);
    const [showDropdown, setShowDropdown] = useState(false);

    // Reviews
    const [reviews, setReviews]           = useState([]);
    const [reviewsLoading, setReviewsLoading] = useState(true);
    const [reviewError, setReviewError]   = useState("");
    const [submitMsg, setSubmitMsg]       = useState(null);
    const [submitting, setSubmitting]     = useState(false);
    const [newRating, setNewRating]       = useState(0);
    const [hoveredStar, setHoveredStar]   = useState(0);
    const [newComment, setNewComment]     = useState("");

    useEffect(() => {
        const token = localStorage.getItem("token");
        const user  = localStorage.getItem("user");
        if (token) {
            setIsLoggedIn(true);
            if (user) setUserData(JSON.parse(user));
        }
    }, []);

    const handleLogout = () => {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        localStorage.removeItem("cart");
        setIsLoggedIn(false);
        setShowDropdown(false);
        setUserData(null);
    };

    // Fetch single product
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

    // Fetch reviews
    useEffect(() => {
        const fetchReviews = async () => {
            try {
                const res = await fetch(`${API_URL}/api/products/${id}/reviews`);
                if (!res.ok) {
                    setReviews([]);
                    return;
                }
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

    // Submit review
    const handleSubmitReview = async () => {
        if (newRating === 0) { setReviewError("Please select a star rating."); return; }
        if (!newComment.trim()) { setReviewError("Please write a comment."); return; }

        setReviewError("");
        setSubmitting(true);

        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_URL}/api/reviews`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${token}`,
                },
                body: JSON.stringify({ product_id: id, rating: newRating, comment: newComment }),
            });

            if (!res.ok) {
                let errMsg = "Failed to submit review.";
                try {
                    const data = await res.json();
                    errMsg = data.error || errMsg;
                } catch {}
                setReviewError(errMsg);
                return;
            }

            // Refresh reviews list
            const refreshed = await fetch(`${API_URL}/api/products/${id}/reviews`);
            const refreshedData = await refreshed.json();
            setReviews(refreshedData || []);

            // Refresh product to get updated rating/review_count
            const productRes = await fetch(`${API_URL}/api/products/${id}`);
            const productData = await productRes.json();
            setProduct(productData);

            setNewRating(0);
            setNewComment("");
            setSubmitMsg("Review submitted! It will appear after moderation.");
            setTimeout(() => setSubmitMsg(null), 3000);
        } catch (err) {
            console.error(err);
            setReviewError("Server error. Please try again.");
        } finally {
            setSubmitting(false);
        }
    };

    // Add to Cart (UPDATED FOR PERSISTENT CART)
    const handleAddToCart = async () => {
        if (!product || product.quantity === 0) return;

        if (isLoggedIn) {
            // --- LOGGED-IN USER: Send to Backend ---
            try {
                const token = localStorage.getItem("token");
                const response = await fetch(`${API_URL}/api/cart/item`, {
                    method: 'PATCH',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${token}` 
                    },
                    body: JSON.stringify({
                        product_id: product.id,
                        quantity: quantity
                    })
                });

                if (response.ok) {
                    setCartMsg("added");
                    setTimeout(() => setCartMsg(null), 1500);
                } else {
                    const errorData = await response.json();
                    console.error("Failed to add to backend cart:", errorData.error);
                    setCartMsg("maxed"); 
                    setTimeout(() => setCartMsg(null), 1500);
                }
            } catch (error) {
                console.error("Network error adding to cart:", error);
            }
        } else {
            // --- GUEST USER: Save to LocalStorage ---
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

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">

            {/* Header */}
            <header className="w-full bg-white shadow-md relative z-50">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4 flex items-center gap-4">
                    <Link to="/" className="text-xl sm:text-2xl font-bold text-gray-800 flex-shrink-0">
                        MyStore
                    </Link>
                    <div className="flex-1" />
                    <nav className="flex items-center space-x-3 sm:space-x-6 flex-shrink-0">
                        <Link to="/shoppingcart" className="text-gray-700 hover:text-blue-600 text-sm sm:text-base">
                            🛒 <span className="hidden sm:inline">Cart</span>
                        </Link>
                        {isLoggedIn ? (
                            <div className="relative">
                                <button
                                    onClick={() => setShowDropdown(!showDropdown)}
                                    className="flex items-center space-x-1 text-gray-700 hover:text-blue-600 text-sm sm:text-base"
                                >
                                    <span className="font-medium">{userData?.name || "My Account"}</span>
                                    <span className="text-xs">▾</span>
                                </button>
                                {showDropdown && (
                                    <div className="absolute right-0 mt-3 w-48 bg-white border rounded-lg shadow-lg z-50">
                                        <Link to="/profile" className="block px-4 py-2 hover:bg-gray-100 text-sm">Profile</Link>
                                        <Link to="/orders"  className="block px-4 py-2 hover:bg-gray-100 text-sm">My Orders</Link>
                                        <button onClick={handleLogout} className="block w-full text-left px-4 py-2 text-red-600 hover:bg-gray-100 text-sm">
                                            Log Out
                                        </button>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <>
                                <Link to="/login"  className="text-gray-700 hover:text-blue-600 text-sm sm:text-base">Login</Link>
                                <Link to="/signup" className="bg-blue-600 text-white px-3 sm:px-5 py-2 rounded-lg hover:bg-blue-700 text-sm sm:text-base">Sign Up</Link>
                            </>
                        )}
                    </nav>
                </div>
            </header>

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
                                    {product.rating} ({product.review_count} reviews)
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

                {/* ── Reviews Section ── */}
                <div className="bg-white rounded-2xl shadow-md p-6 sm:p-8 flex flex-col gap-6">
                    <h2 className="text-xl font-bold text-gray-900">
                        Customer Reviews
                        {reviews.length > 0 && (
                            <span className="ml-2 text-sm font-normal text-gray-400">({reviews.length})</span>
                        )}
                    </h2>

                    {/* Submit Review Form — only for logged-in users */}
                    {isLoggedIn ? (
                        <div className="border border-gray-200 rounded-xl p-5 flex flex-col gap-4 bg-gray-50">
                            <h3 className="text-sm font-semibold text-gray-700">Write a Review</h3>

                            {/* Star Picker */}
                            <div className="flex items-center gap-1">
                                {[1, 2, 3, 4, 5].map((star) => (
                                    <button
                                        key={star}
                                        type="button"
                                        onMouseEnter={() => setHoveredStar(star)}
                                        onMouseLeave={() => setHoveredStar(0)}
                                        onClick={() => setNewRating(star)}
                                        className="text-2xl transition-transform hover:scale-110 focus:outline-none"
                                    >
                                        <span className={(hoveredStar || newRating) >= star ? "text-yellow-400" : "text-gray-300"}>
                                            ★
                                        </span>
                                    </button>
                                ))}
                                {newRating > 0 && (
                                    <span className="ml-2 text-sm text-gray-500">
                                        {["", "Poor", "Fair", "Good", "Very Good", "Excellent"][newRating]}
                                    </span>
                                )}
                            </div>

                            {/* Comment */}
                            <textarea
                                value={newComment}
                                onChange={(e) => setNewComment(e.target.value)}
                                placeholder="Share your experience with this product..."
                                rows={3}
                                className="w-full border border-gray-300 rounded-lg px-4 py-3 text-sm text-gray-700 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            />

                            {reviewError && (
                                <p className="text-sm text-red-600">{reviewError}</p>
                            )}
                            {submitMsg && (
                                <p className="text-sm text-green-600">✓ {submitMsg}</p>
                            )}

                            <button
                                onClick={handleSubmitReview}
                                disabled={submitting}
                                className="self-end px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {submitting ? "Submitting..." : "Submit Review"}
                            </button>
                        </div>
                    ) : (
                        <div className="border border-dashed border-gray-300 rounded-xl p-5 text-center text-sm text-gray-500">
                            <Link to="/login" className="text-blue-600 hover:underline font-medium">Log in</Link>
                            {" "}to leave a review.
                        </div>
                    )}

                    {/* Reviews List */}
                    {reviewsLoading ? (
                        <p className="text-sm text-gray-400">Loading reviews...</p>
                    ) : reviews.length === 0 ? (
                        <p className="text-sm text-gray-400">No reviews yet. Be the first to review this product!</p>
                    ) : (
                        <div className="flex flex-col divide-y divide-gray-100">
                            {reviews.map((review) => (
                                <div key={review.id} className="py-5 flex flex-col gap-2">
                                    <div className="flex items-center justify-between">
                                        <div className="flex items-center gap-3">
                                            {/* Avatar */}
                                            <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold text-sm flex-shrink-0">
                                                {review.user_name?.charAt(0).toUpperCase() || "?"}
                                            </div>
                                            <div>
                                                <p className="text-sm font-semibold text-gray-800">{review.user_name}</p>
                                                <p className="text-xs text-gray-400">{formatDate(review.created_at)}</p>
                                            </div>
                                        </div>
                                        {renderStars(review.rating, "text-base")}
                                    </div>
                                    {review.comment && (
                                        <p className="text-sm text-gray-600 leading-relaxed pl-11">
                                            {review.comment}
                                        </p>
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