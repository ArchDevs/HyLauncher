import React, { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ExternalLink, Calendar } from 'lucide-react';
import { GetNews } from '../../wailsjs/go/app/App';
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime';

interface BlogPost {
    _id: string;
    title: string;
    publishedAt: string;
    slug: string;
    coverImage: {
        s3Key: string;
    };
    bodyExcerpt: string;
    author: string;
}

export const NewsSection: React.FC = () => {
    const [news, setNews] = useState<BlogPost[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchNews = async () => {
            try {
                // @ts-ignore
                const posts = await GetNews();
                if (posts && posts.length > 0) {
                    setNews(posts);
                }
            } catch (err) {
                console.error("Failed to fetch news:", err);
            } finally {
                setLoading(false);
            }
        };
        fetchNews();
    }, []);

    // Helper to format nice dates
    const formatDate = (dateStr: string) => {
        const date = new Date(dateStr);
        return new Intl.DateTimeFormat('en-US', { month: 'short', day: 'numeric', year: 'numeric' }).format(date);
    };

    // Helper to construct URL
    // Format: https://hytale.com/news/2026/01/slug
    const getPostUrl = (post: BlogPost) => {
        if (!post) return 'https://hytale.com/news';
        const date = new Date(post.publishedAt);
        const year = date.getFullYear();
        // Month is 0-indexed, padStart ensures '01' instead of '1'
        const month = String(date.getMonth() + 1).padStart(2, '0');
        return `https://hytale.com/news/${year}/${month}/${post.slug}`;
    };

    if (loading) {
        return (
            <div className="w-[532px] h-[320px] bg-[#090909]/[0.55] backdrop-blur-xl rounded-[14px] border border-[#FFA845]/[0.10] p-4 flex items-center justify-center">
                <div className="flex flex-col items-center gap-2">
                    <div className="w-5 h-5 border-2 border-orange-500/30 border-t-orange-500 rounded-full animate-spin" />
                    <span className="text-xs text-gray-500 font-mono uppercase tracking-widest">Loading Feed</span>
                </div>
            </div>
        );
    }

    if (news.length === 0) {
        return (
            <div className="w-[532px] h-[320px] bg-[#090909]/[0.55] backdrop-blur-xl rounded-[14px] border border-[#FFA845]/[0.10] p-4 flex items-center justify-center">
                <span className="text-sm text-gray-500 italic">No news available at the moment.</span>
            </div>
        );
    }

    return (
        <div className="relative w-[532px] h-[320px] bg-[#090909]/[0.55] backdrop-blur-xl rounded-[20px] border border-[#FFA845]/[0.10] overflow-hidden group flex flex-col">
            {/* Header */}
            <div className="px-5 py-4 border-b border-white/5 flex items-center justify-between bg-white/[0.02]">
                <h3 className="text-xs font-black text-gray-400 uppercase tracking-[0.2em] flex items-center gap-2">
                    <div className="w-1 h-1 bg-orange-500 rounded-full animate-pulse" />
                    Latest News
                </h3>
                <button
                    onClick={() => BrowserOpenURL("https://hytale.com/news")}
                    className="text-[10px] font-bold text-gray-500 hover:text-orange-400 transition-colors flex items-center gap-1.5 cursor-pointer"
                >
                    BROWSE ALL <ExternalLink size={10} />
                </button>
            </div>

            {/* Scrollable List */}
            <div className="flex-1 overflow-y-auto custom-scrollbar p-2">
                <div className="flex flex-col gap-2">
                    {news.map((post, idx) => (
                        <motion.div
                            key={post._id}
                            initial={{ opacity: 0, x: -10 }}
                            animate={{ opacity: 1, x: 0 }}
                            transition={{ delay: idx * 0.1 }}
                            className="bg-white/5 hover:bg-white/[0.08] border border-white/5 hover:border-[#FFA845]/20 rounded-[12px] p-3 transition-all duration-300 group/item cursor-pointer"
                            onClick={() => BrowserOpenURL(getPostUrl(post))}
                        >
                            <div className="flex gap-4">
                                <div className="w-[100px] aspect-[16/9] rounded-lg overflow-hidden border border-white/5 shrink-0">
                                    <img
                                        src={`https://cdn.hytale.com/variants/blog_thumb_${post.coverImage.s3Key}`}
                                        alt=""
                                        className="w-full h-full object-cover group-hover/item:scale-110 transition-transform duration-500"
                                    />
                                </div>
                                <div className="flex flex-col justify-center min-w-0">
                                    <div className="flex items-center gap-2 text-[9px] font-bold text-orange-400/70 uppercase tracking-tighter mb-1">
                                        <Calendar size={10} />
                                        <span>{formatDate(post.publishedAt)}</span>
                                    </div>
                                    <h4 className="text-sm font-bold text-gray-100 line-clamp-1 group-hover/item:text-orange-400 transition-colors">
                                        {post.title}
                                    </h4>
                                    <p className="text-[11px] text-gray-400 line-clamp-1 opacity-60 mt-1">
                                        {post.bodyExcerpt}
                                    </p>
                                </div>
                            </div>
                        </motion.div>
                    ))}
                </div>
            </div>

            <style>{`
                .custom-scrollbar::-webkit-scrollbar {
                    width: 4px;
                }
                .custom-scrollbar::-webkit-scrollbar-track {
                    background: transparent;
                }
                .custom-scrollbar::-webkit-scrollbar-thumb {
                    background: rgba(255, 168, 69, 0.1);
                    border-radius: 10px;
                }
                .custom-scrollbar::-webkit-scrollbar-thumb:hover {
                    background: rgba(255, 168, 69, 0.3);
                }
            `}</style>
        </div>
    );
};
