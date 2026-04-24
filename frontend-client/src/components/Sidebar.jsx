import { Link } from "react-router-dom";

export default function Sidebar() {
  return (
    <aside className="hidden lg:flex flex-col w-64 h-screen sticky top-0 bg-[#1c110b] border-r border-[#40322a] p-6">

      <div className="mb-10">
        <h2 className="text-xl font-serif text-[#e7b4ff]">
          The Archivist
        </h2>
        <p className="text-xs text-[#9a8c9b] uppercase tracking-widest">
          Master Merchant
        </p>
      </div>

      <nav className="flex flex-col gap-2 text-sm">

        <Link to="/"className="px-4 py-2 bg-[#342720] text-[#e7b4ff] rounded-lg">
          Shop
        </Link>

        <Link
          to="/orders"
          className="px-4 py-2 text-[#d1c5b0] hover:bg-[#342720] rounded-lg"
        >
          Orders
        </Link>

        <Link
          to="/profile"
          className="px-4 py-2 text-[#d1c5b0] hover:bg-[#342720] rounded-lg"
        >
          Profile
        </Link>

        <Link
          to="/shoppingcart"
          className="px-4 py-2 text-[#d1c5b0] hover:bg-[#342720] rounded-lg"
        >
          Cart
        </Link>

      </nav>

    </aside>
  );
}
