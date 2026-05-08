import { useState, useEffect } from "react";

const API_BASE = import.meta.env.VITE_API_URL;

const RATING_LABELS = ["", "Poor", "Fair", "Good", "Very Good", "Excellent"];

function StarDisplay({ rating }) {
    if (!rating) return null;
    return (
        <div className="flex items-center gap-1">
            <span className="text-yellow-400 text-sm">
                {"★".repeat(rating)}{"☆".repeat(5 - rating)}
            </span>
            <span className="text-xs text-gray-400">{RATING_LABELS[rating]}</span>
        </div>
    );
}

function ReviewCard({ review, rating, onModerated }) {
    const [saving, setSaving]     = useState(null);
    const [feedback, setFeedback] = useState(null);

    const date = new Date(review.created_at).toLocaleDateString("en-US", {
        year: "numeric", month: "short", day: "numeric",
    });

    async function handleAction(action) {
        setSaving(action);
        setFeedback(null);
        try {
            const token = localStorage.getItem("token");
            const res = await fetch(`${API_BASE}/api/reviews/${review.id}/moderate`, {
                method: "PATCH",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": `Bearer ${token}`,
                },
                body: JSON.stringify({ action }),
            });
            if (!res.ok) {
                const err = await res.json();
                throw new Error(err.error || "Action failed");
            }
            onModerated(review.id);
        } catch (err) {
            setFeedback(err.message);
            setSaving(null);
        }
    }

    return (
        <div className="bg-white border border-gray-200 rounded-xl overflow-hidden">

            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 px-5 py-4">
                <div className="flex items-center gap-3">
                    <div className="w-9 h-9 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold text-sm shrink-0">
                        {review.user_name?.charAt(0).toUpperCase() || "?"}
                    </div>
                    <div>
                        <p className="text-sm font-semibold text-gray-900">{review.user_name}</p>
                        <p className="text-xs text-gray-400">{date}</p>
                    </div>
                </div>
                <StarDisplay rating={rating} />
            </div>

            {/* Product ID */}
            <div className="px-5 pb-3 border-t border-gray-100 pt-3">
                <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide">Product ID</span>
                <p className="text-xs text-gray-600 font-mono mt-0.5">{review.product_id}</p>
            </div>

            {/* Comment */}
            <div className="px-5 pb-4">
                <span className="text-xs font-semibold text-gray-400 uppercase tracking-wide">Comment</span>
                <p className="text-sm text-gray-700 mt-1 leading-relaxed">
                    {review.comment || <em className="text-gray-400">No comment</em>}
                </p>
            </div>

            {/* Actions */}
            <div className="px-5 py-4 border-t border-gray-100 bg-gray-50 flex items-center gap-3">
                <button
                    onClick={() => handleAction("approve")}
                    disabled={!!saving}
                    className="flex-1 bg-green-600 hover:bg-green-700 disabled:bg-gray-300 disabled:cursor-not-allowed text-white text-sm font-semibold py-2 rounded-lg transition-colors"
                >
                    {saving === "approve" ? "Approving…" : "✓ Approve"}
                </button>
                <button
                    onClick={() => handleAction("reject")}
                    disabled={!!saving}
                    className="flex-1 bg-red-500 hover:bg-red-600 disabled:bg-gray-300 disabled:cursor-not-allowed text-white text-sm font-semibold py-2 rounded-lg transition-colors"
                >
                    {saving === "reject" ? "Rejecting…" : "✕ Reject"}
                </button>
            </div>

            {feedback && (
                <div className="px-5 py-2.5 text-sm font-medium border-t bg-red-50 text-red-700 border-red-100">
                    {feedback}
                </div>
            )}
        </div>
    );
}

export default function ReviewManager() {
    const [reviews, setReviews]       = useState([]);
    const [loading, setLoading]       = useState(true);
    const [error, setError]           = useState(null);
    // Map of "productId:userId" -> rating number
    const [ratingsMap, setRatingsMap] = useState({});

    useEffect(() => {
        async function fetchPending() {
            try {
                const token = localStorage.getItem("token");
                const res = await fetch(`${API_BASE}/api/reviews/pending`, {
                    headers: { "Authorization": `Bearer ${token}` },
                });
                if (!res.ok) throw new Error("Failed to fetch pending reviews");
                const data = await res.json();
                const pending = data || [];
                setReviews(pending);

                // Collect unique product IDs and fetch their ratings
                const uniqueProductIds = [...new Set(pending.map((r) => r.product_id))];
                const map = {};
                await Promise.all(
                    uniqueProductIds.map(async (productId) => {
                        try {
                            const rRes = await fetch(`${API_BASE}/api/products/${productId}/ratings`);
                            if (!rRes.ok) return;
                            const ratings = await rRes.json();
                            (ratings || []).forEach((r) => {
                                map[`${productId}:${r.user_id}`] = r.rating;
                            });
                        } catch { /* skip */ }
                    })
                );
                setRatingsMap(map);
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        }
        fetchPending();
    }, []);

    function handleModerated(reviewId) {
        setReviews((prev) => prev.filter((r) => r.id !== reviewId));
    }

    if (loading) {
        return (
            <div className="flex items-center justify-center py-24">
                <div className="flex flex-col items-center gap-3">
                    <div className="w-8 h-8 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin" />
                    <p className="text-sm text-gray-500">Loading pending reviews…</p>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-2xl mx-auto">
            <div className="flex items-center justify-between mb-4">
                <h1 className="text-2xl font-bold text-gray-900">Review Moderation</h1>
                {reviews.length > 0 && (
                    <span className="text-sm text-gray-400">{reviews.length} pending</span>
                )}
            </div>

            {error && (
                <div className="bg-red-50 border border-red-200 text-red-700 text-sm font-medium rounded-lg px-4 py-3 mb-4">
                    {error}
                </div>
            )}

            {!error && reviews.length === 0 && (
                <div className="bg-white border border-gray-200 rounded-xl px-6 py-14 text-center">
                    <p className="text-4xl mb-3">✅</p>
                    <p className="text-gray-700 font-semibold">All caught up!</p>
                    <p className="text-sm text-gray-400 mt-1">No reviews are pending moderation.</p>
                </div>
            )}

            <div className="flex flex-col gap-3">
                {reviews.map((review) => (
                    <ReviewCard
                        key={review.id}
                        review={review}
                        rating={ratingsMap[`${review.product_id}:${review.user_id}`]}
                        onModerated={handleModerated}
                    />
                ))}
            </div>
        </div>
    );
}