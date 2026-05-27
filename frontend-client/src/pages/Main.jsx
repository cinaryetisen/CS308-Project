import { Link, useNavigate, useOutletContext, useSearchParams } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { apiRequest } from '../api/client';

export default function Main() {
    const navigate = useNavigate();

    const { refreshCartCount } = useOutletContext();

    const [searchParams] = useSearchParams();
    const categoryParam = searchParams.get("category");

    const [products, setProducts]         = useState([]);
    const [loading, setLoading]           = useState(true);
    const [sortOption, setSortOption]     = useState("default");
    const [searchQuery, setSearchQuery]   = useState("");
    const [debouncedSearch, setDebouncedSearch] = useState("");
    const [cartFeedback, setCartFeedback] = useState({});

    useEffect(() => {
        const timer = setTimeout(() => setDebouncedSearch(searchQuery), 300);
        return () => clearTimeout(timer);
    }, [searchQuery]);

    const [isLoggedIn, setIsLoggedIn] = useState(false);
    useEffect(() => {
        if (localStorage.getItem("token")) setIsLoggedIn(true);
    }, []);

    useEffect(() => {
        const fetchProducts = async () => {
            setLoading(true);
            try {
                let path = "/api/products?";
                if (debouncedSearch) path += `search=${encodeURIComponent(debouncedSearch)}&`;
                if (categoryParam)   path += `category=${encodeURIComponent(categoryParam)}&`;
                if (sortOption !== "default") {
                    const sortMap = { "price-asc": "price_asc", "price-desc": "price_desc", "rating-desc": "popular" };
                    path += `sort=${sortMap[sortOption]}`;
                }
                const data = await apiRequest(path, {}, false);
                setProducts(Array.isArray(data) ? data : []);
            } catch (error) {
                console.error("Error fetching products:", error);
            } finally {
                setLoading(false);
            }
        };
        fetchProducts();
    }, [debouncedSearch, sortOption, categoryParam]);

    const addToCart = async (e, product) => {
        e.stopPropagation();
        if (product.quantity === 0) return;
        if (isLoggedIn) {
            try {
                await apiRequest("/api/cart/item", {
                    method: "PATCH",
                    body: JSON.stringify({ product_id: product.id, quantity: 1 })
                });
                setCartFeedback((prev) => ({ ...prev, [product.id]: "added" }));
                setTimeout(() => setCartFeedback((prev) => ({ ...prev, [product.id]: null })), 1200);
                refreshCartCount();
            } catch (error) {
                console.error("Failed to add to cart:", error);
                setCartFeedback((prev) => ({ ...prev, [product.id]: "error" }));
                setTimeout(() => setCartFeedback((prev) => ({ ...prev, [product.id]: null })), 2000);
            }
        } else {
            const cart = JSON.parse(localStorage.getItem("cart") || "[]");
            const idx  = cart.findIndex((item) => item.id === product.id);
            if (idx >= 0) {
                if (cart[idx].quantity >= product.quantity) {
                    setCartFeedback((prev) => ({ ...prev, [product.id]: "maxed" }));
                    setTimeout(() => setCartFeedback((prev) => ({ ...prev, [product.id]: null })), 1500);
                    return;
                }
                cart[idx].quantity = Math.min(cart[idx].quantity + 1, product.quantity);
            } else {
                cart.push({ id: product.id, name: product.name, price: product.price, image_url: product.image_url, stock: product.quantity, quantity: 1 });
            }
            localStorage.setItem("cart", JSON.stringify(cart));
            refreshCartCount();
            setCartFeedback((prev) => ({ ...prev, [product.id]: "added" }));
            setTimeout(() => setCartFeedback((prev) => ({ ...prev, [product.id]: null })), 1200);
        }
    };

    return (
        <div className="min-h-screen bg-[#1a0f0a] flex flex-col">
            <main className="flex-1 max-w-7xl mx-auto w-full px-4 sm:px-6 py-8 sm:py-10">
                <div className="flex flex-col md:flex-row justify-between items-start md:items-center mb-8 gap-4">
                    <h2 className="text-4xl font-serif text-[#f5ded3] capitalize">
                        {categoryParam ? `The Vault — ${categoryParam}` : "The Vault of Essence"}
                    </h2>
                    <div className="flex flex-col sm:flex-row gap-4 w-full md:w-auto">
                        <input
                            type="text"
                            placeholder="Search artifacts..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full sm:w-64 bg-[#251912] border border-[#342720] px-4 py-2 rounded text-sm text-[#f5ded3] placeholder-[#9a8c9b] focus:outline-none focus:border-[#e7b4ff] transition-colors"
                        />
                        <select
                            value={sortOption}
                            onChange={(e) => setSortOption(e.target.value)}
                            className="w-full sm:w-auto bg-[#251912] border border-[#342720] px-3 py-2 rounded text-sm text-[#f5ded3] focus:outline-none focus:border-[#e7b4ff] transition-colors"
                        >
                            <option value="default">Featured</option>
                            <option value="price-asc">Price ↑</option>
                            <option value="price-desc">Price ↓</option>
                            <option value="rating-desc">Popularity</option>
                        </select>
                    </div>
                </div>

                {loading ? (
                    <p className="text-center py-20 text-[#9a8c9b]">Loading...</p>
                ) : products.length === 0 ? (
                    <p className="text-center py-20 text-[#9a8c9b]">No products found.</p>
                ) : (
                    <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-8">
                        {products.map((product) => {
                            const outOfStock      = product.quantity === 0;
                            const feedback        = cartFeedback[product.id];
                            const discountedPrice = product.discount > 0 ? product.price * (1 - product.discount / 100) : null;

                            return (
                                <div
                                    key={product.id}
                                    onClick={() => navigate(`/products/${product.id}`)}
                                    className="group relative overflow-hidden bg-[#251912] border border-[#342720] hover:border-[#8a47af] hover:shadow-[0_0_20px_rgba(138,71,175,0.15)] rounded-lg p-4 transition-all duration-300 cursor-pointer flex flex-col"
                                >
                                    <div className="absolute inset-0 bg-gradient-to-r from-[#e7b4ff]/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none" />

                                    <div className="relative z-10 aspect-[4/5] mb-4 overflow-hidden rounded border border-[#342720]/50">
                                        <img src={product.image_url} alt={product.name} className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110" />
                                        <div className={`absolute top-3 right-3 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest shadow-sm ${outOfStock ? 'bg-[#93000a] text-[#ffdad6]' : 'bg-[#add461] text-[#131f00]'}`}>
                                            {outOfStock ? "Out of Stock" : "In Stock"}
                                        </div>
                                        {discountedPrice && (
                                            <div className="absolute top-3 left-3 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-widest bg-[#e7b4ff] text-[#300049] shadow-sm">
                                                -{product.discount}%
                                            </div>
                                        )}
                                    </div>

                                    <span className="relative z-10 text-[10px] uppercase tracking-widest text-[#add461] mb-1">{product.category || "Artifact"}</span>
                                    <h3 className="relative z-10 text-lg font-serif text-[#f5ded3] line-clamp-2">{product.name}</h3>

                                    {/* Rating — 1 decimal place */}
                                    <div className="relative z-10 text-sm text-[#9a8c9b] mt-1">
                                        <span className="text-[#add461]">⭐ {Number(product.rating).toFixed(1)}</span> ({product.review_count} reviews)
                                    </div>

                                    <div className="relative z-10 mt-2 mb-4">
                                        {discountedPrice ? (
                                            <div className="flex items-center gap-2">
                                                <span className="text-[#e7b4ff] font-bold">${discountedPrice.toFixed(2)}</span>
                                                <span className="text-sm text-[#9a8c9b] line-through">${Number(product.price).toFixed(2)}</span>
                                            </div>
                                        ) : (
                                            <span className="text-[#e7b4ff] font-bold">${Number(product.price).toFixed(2)}</span>
                                        )}
                                    </div>

                                    <button
                                        onClick={(e) => addToCart(e, product)}
                                        disabled={outOfStock}
                                        className={`relative z-10 mt-auto w-full py-2 rounded font-semibold active:scale-95 transition-all duration-150 shadow-md ${
                                            outOfStock ? "bg-[#342720] text-[#9a8c9b] cursor-not-allowed"
                                                : feedback === "added" ? "bg-[#add461] text-[#131f00]"
                                                    : feedback === "maxed" ? "bg-[#93000a] text-[#ffdad6]"
                                                        : feedback === "error" ? "bg-[#93000a] text-[#ffdad6]"
                                                            : "bg-gradient-to-r from-[#e7b4ff] to-[#8a47af] text-[#300049] hover:brightness-110"
                                        }`}
                                    >
                                        {outOfStock ? "Unavailable" : feedback === "added" ? "✓ Added" : feedback === "maxed" ? "Max Reached" : feedback === "error" ? "Failed" : "Add to Cart"}
                                    </button>
                                </div>
                            );
                        })}
                    </div>
                )}
            </main>
        </div>
    );
}