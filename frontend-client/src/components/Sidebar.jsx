import { Link } from "react-router-dom";
import { useState, useEffect } from "react";
import { apiRequest } from "../api/client";

export default function Sidebar() {
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchCategoriesFromProducts = async () => {
      try {
        const products = await apiRequest("/api/products", {}, false);
        const allCategories = products.map(product => product.category).filter(Boolean);
        setCategories([...new Set(allCategories)]);
      } catch (error) {
        console.error("Error fetching categories:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchCategoriesFromProducts();
  }, []);

  return (
    // 👇 Removed h-screen and sticky. Added h-full and overflow-y-auto.
    <aside className="hidden lg:flex flex-col w-64 h-full bg-[#1c110b] border-r border-[#40322a] p-6 z-40 overflow-y-auto">
      <div className="mb-10 shrink-0">
        <h2 className="text-xl font-serif text-[#e7b4ff]">Categories</h2>
        <p className="text-xs text-[#9a8c9b] uppercase tracking-widest mt-1">
          Filter Wares
        </p>
      </div>

      <nav className="flex flex-col gap-2 text-sm pb-6">
        <Link
          to="/"
          className="px-4 py-2 text-[#e7b4ff] hover:bg-[#342720] rounded-lg transition mb-2 shrink-0"
        >
          All Items
        </Link>

        {loading ? (
          <div className="px-4 py-2 text-[#9a8c9b] animate-pulse">
            Consulting the archives...
          </div>
        ) : categories.length === 0 ? (
          <div className="px-4 py-2 text-[#9a8c9b] italic">
            No categories found.
          </div>
        ) : (
          categories.map((category) => (
            <Link
              key={category} 
              to={`/?category=${encodeURIComponent(category)}`}
              className="px-4 py-2 text-[#d1c5b0] hover:bg-[#342720] hover:text-[#e7b4ff] rounded-lg transition capitalize shrink-0"
            >
              {category}
            </Link>
          ))
        )}
      </nav>
    </aside>
  );
}