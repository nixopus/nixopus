"use client";

import { ChevronRight, Star } from "lucide-react";
import Link from "next/link";
import { motion } from "framer-motion";

export function Banner() {
  return (
    <motion.div 
      initial={{ opacity: 0, y: -20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className="relative top-0 bg-gradient-to-r from-indigo-500/10 to-purple-500/10 py-3 text-white md:py-0 border-b border-white/10 backdrop-blur-sm"
    >
      <div className="container flex flex-col items-center justify-center gap-4 md:h-12 md:flex-row">
        <Link
          href="https://github.com/raghavyuva/nixopus"
          target="_blank"
          rel="noopener noreferrer"
          className="group inline-flex items-center justify-center text-center text-sm leading-loose hover:text-indigo-300 transition-colors duration-300"
        >
          <span className="flex items-center gap-2">
            <Star className="size-4 text-yellow-400 animate-pulse" />
            <span className="font-medium">
              Star us on GitHub to support our open-source journey
            </span>
            <ChevronRight className="ml-1 size-4 transition-all duration-300 ease-out group-hover:translate-x-1" />
          </span>
        </Link>
      </div>
    </motion.div>
  );
}