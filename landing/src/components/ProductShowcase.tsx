"use client"
import appScreen from "../assets/images/hosted_apps.jpeg";
import Image from 'next/image';
import { motion, useScroll, useTransform } from 'framer-motion';
import { useRef } from "react";
import { Sparkles } from 'lucide-react';

export const ProductShowcase = () => {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start end", "end end"]
  });

  const rotateX = useTransform(scrollYProgress, [0, 1], [15, 0]);
  const opacity = useTransform(scrollYProgress, [0, 0.5, 1], [0.3, 0.8, 1]);
  const scale = useTransform(scrollYProgress, [0, 1], [0.95, 1]);

  return (
    <section 
      ref={containerRef}
      className="relative overflow-hidden py-32"
    >
      <div className="absolute inset-0 bg-gradient-to-b from-gray-900 via-gray-950 to-black"></div>
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_120%,rgba(120,119,198,0.1),rgba(255,255,255,0))]"></div>
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
      
      <div className="container relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          viewport={{ once: true }}
          className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-indigo-500/10 to-purple-500/10 mb-6 mx-auto"
        >
          <Sparkles className="size-4 text-yellow-400" />
          <span className="text-sm font-medium text-white/70">Product Showcase</span>
        </motion.div>

        <motion.h2 
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.2 }}
          viewport={{ once: true }}
          className="text-4xl md:text-5xl font-bold tracking-tight text-center text-white mb-6"
        >
          Intuitive <span className="bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">Interface</span>
        </motion.h2>

        <motion.p 
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.3 }}
          viewport={{ once: true }}
          className="text-lg text-white/70 text-center max-w-2xl mx-auto mb-16"
        >
          Experience a seamless and intuitive interface designed for modern infrastructure management. 
          Everything you need, right at your fingertips.
        </motion.p>

        <motion.div
          style={{
            opacity,
            rotateX,
            scale,
            transformPerspective: "1000px",
          }}
          className="relative"
        >
          <div className="absolute inset-0 bg-gradient-to-r from-indigo-500/20 to-purple-500/20 rounded-2xl blur-2xl -z-10"></div>
          <Image 
            src={appScreen} 
            alt="Nixopus interface showcase" 
            className="rounded-2xl shadow-2xl shadow-indigo-500/20"
          />
        </motion.div>
      </div>
    </section>
  )
};
