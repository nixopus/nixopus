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
  const opacity = useTransform(scrollYProgress, [0, 0.5, 1], [0, 1, 1]);

  return (
    <section 
      ref={containerRef} 
      className={cn(
        "relative overflow-hidden py-32",
        "bg-gradient-to-b from-gray-900 via-gray-950 to-black"
      )}
    >
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_120%,rgba(120,119,198,0.1),rgba(255,255,255,0))]"></div>
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
      
      <motion.div 
        style={{ y: translateY, opacity }} 
        className="relative container max-w-4xl text-center"
      >
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          viewport={{ once: true }}
          className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-indigo-500/10 to-purple-500/10 mb-6"
        >
          <Star className="size-4 text-yellow-400" />
          <span className="text-sm font-medium text-white/70">Open Source</span>
        </motion.div>

        <motion.h2 
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.2 }}
          viewport={{ once: true }}
          className="text-4xl md:text-5xl font-bold tracking-tight text-white mb-6"
        >
          We love <span className="bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">open-source</span>
        </motion.h2>

        <motion.p 
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.3 }}
          viewport={{ once: true }}
          className="text-lg text-white/70 max-w-2xl mx-auto mb-10"
        >
          Our source code is available on GitHub - feel free to read, review, or contribute to it
          however you want!
        </motion.p>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.4 }}
          viewport={{ once: true }}
          className="flex justify-center"
        >
          <Link
            href="https://github.com/raghavyuva/nixopus"
            target="_blank"
            rel="noopener noreferrer"
            className="group inline-flex items-center gap-2 rounded-lg bg-gradient-to-r from-indigo-500 to-purple-500 px-6 py-3 text-white transition-all duration-300 hover:from-indigo-600 hover:to-purple-600 hover:shadow-lg hover:shadow-indigo-500/20"
          >
            <GithubIcon className="size-5" />
            <span>Star us on GitHub</span>
            <div className="ml-2 flex items-center gap-1 rounded-full bg-white/10 px-3 py-1 text-sm">
              <Star className="size-4" />
              <span>9999</span>
            </div>
          </Link>
        </motion.div>
      </motion.div>
    </section>
  )
};