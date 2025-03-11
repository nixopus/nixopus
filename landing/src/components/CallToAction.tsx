"use client"
import { motion, useScroll, useTransform } from 'framer-motion';
import { useRef } from 'react';
import Link from 'next/link';
import { cn } from './lib/utils';
import { GithubIcon, Star } from 'lucide-react';

export const CallToAction = () => {
  const containerRef = useRef<HTMLDivElement>(null);

  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start end", "end end"]
  })

  const translateY = useTransform(scrollYProgress, [0, 1], [50, -50]);

  return (
    <section ref={containerRef} className={cn("mx-auto max-w-2xl px-6 text-center bg-black text-white p-20 relative overflow-hidden")}>
      <motion.div style={{ y: translateY }} className="relative">
        <h2 className="mx-auto mt-8 max-w-2xl text-3xl font-bold tracking-tighter lg:text-5xl">
          We love <span className="bg-gradient-to-r from-yellow-400 to-red-400 bg-clip-text text-transparent">open-source</span>
        </h2>
        <p className="mt-4 text-lg text-fg-muted">
          Our source code is available on GitHub - feel free to read, review, or contribute to it
          however you want!
        </p>
        <div className="mt-10 flex justify-center space-x-2">
          <Link
            href={"https://github.com/nixopus/nixopus"}
            target="_blank"
            rel="noreferrer"
            className="group flex"
          >
            <div className="flex  items-center justify-center space-x-2 rounded-md bg-bg-neutral px-4 transition-colors group-hover:bg-bg-neutral-hover">
              <GithubIcon size={18} />
              <span className="truncate">Star us on GitHub</span>
            </div>
            <div className="flex items-center">
             <Star size={18} />
              <div className="flex  items-center rounded-md bg-bg-neutral px-4 font-medium transition-colors group-hover:bg-bg-neutral-hover">
                9999
              </div>
            </div>
          </Link>
        </div>
      </motion.div>
    </section>
  )
};