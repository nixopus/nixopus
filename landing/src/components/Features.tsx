"use client";
import React, { useState } from "react";
import { cn } from "@/lib/utils";
import Image from "next/image";
import { motion } from "framer-motion";
import { Sparkles, Terminal, Server, Shield, Store, Clock } from "lucide-react";

export function FeaturesSectionDemo() {
  const [hoveredImage, setHoveredImage] = useState<string | null>(null);
  const [isYoutubeHovered, setIsYoutubeHovered] = useState(false);

  const features = [
    {
      title: "Powerful Terminal Integration",
      description: "Seamlessly manage your infrastructure with our advanced terminal interface.",
      icon: <Terminal className="size-6" />,
      className: "col-span-1",
    },
    {
      title: "Self-Hosted Solutions",
      description: "Full control over your deployment with our self-hosting capabilities.",
      icon: <Server className="size-6" />,
      className: "col-span-1",
    },
    {
      title: "Advanced Security",
      description: "Enterprise-grade security features to protect your infrastructure.",
      icon: <Shield className="size-6" />,
      className: "col-span-1",
    },
    {
      title: "Marketplace Integration",
      description: "Access a wide range of pre-built solutions and templates.",
      icon: <Store className="size-6" />,
      className: "col-span-1",
    },
    {
      title: "Automated Scheduling",
      description: "Efficiently manage tasks with our powerful cron job system.",
      icon: <Clock className="size-6" />,
      className: "col-span-1",
    },
    {
      title: "Nixopus Features",
      description: "Explore our image gallery and watch our demo video.",
      skeleton: <SkeletonWithYoutube
        setHoveredImage={setHoveredImage}
        hoveredImage={hoveredImage}
        setIsYoutubeHovered={setIsYoutubeHovered}
        isYoutubeHovered={isYoutubeHovered}
      />,
      className: "border-b col-span-1 lg:col-span-2 dark:border-neutral-800",
    }
  ];

  return (
    <section className="relative z-20 py-20 lg:py-40 overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-b from-gray-900 via-gray-950 to-black"></div>
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_120%,rgba(120,119,198,0.1),rgba(255,255,255,0))]"></div>
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
      
      <div className="container relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-indigo-500/10 text-indigo-400 mb-6">
            <Sparkles className="size-4" />
            <span className="text-sm font-medium">Features</span>
          </div>
          <h2 className="text-4xl md:text-5xl font-bold tracking-tight text-white mb-6">
            Packed with <span className="bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">Features</span>
          </h2>
          <p className="text-lg text-white/70 max-w-2xl mx-auto">
            You expect cool features and Nixopus is here to fulfill that expectation.
          </p>
        </motion.div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {features.slice(0, 5).map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
              viewport={{ once: true }}
              className="group relative bg-white/5 backdrop-blur-sm rounded-2xl p-6 hover:bg-white/10 transition-colors duration-300"
            >
              <div className="flex items-center gap-4 mb-4">
                <div className="p-2 rounded-lg bg-indigo-500/10 text-indigo-400 group-hover:bg-indigo-500/20 transition-colors duration-300">
                  {feature.icon}
                </div>
                <h3 className="text-xl font-semibold text-white">{feature.title}</h3>
              </div>
              <p className="text-white/70">{feature.description}</p>
            </motion.div>
          ))}
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.5 }}
          viewport={{ once: true }}
          className="mt-12"
        >
          <FeatureCard className={features[5].className}>
            <FeatureTitle>{features[5].title}</FeatureTitle>
            <FeatureDescription>{features[5].description}</FeatureDescription>
            <div className="h-full w-full">{features[5].skeleton}</div>
          </FeatureCard>
        </motion.div>
      </div>
    </section>
  );
}

const FeatureCard = ({
  children,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) => {
  return (
    <div className={cn(`p-4 sm:p-8 relative overflow-hidden bg-white/5 backdrop-blur-sm rounded-2xl`, className)}>
      {children}
    </div>
  );
};

const FeatureTitle = ({ children }: { children?: React.ReactNode }) => {
  return (
    <p className="max-w-5xl mx-auto text-left tracking-tight text-white text-xl md:text-2xl md:leading-snug font-semibold">
      {children}
    </p>
  );
};

const FeatureDescription = ({ children }: { children?: React.ReactNode }) => {
  return (
    <p className={cn(
      "text-sm md:text-base max-w-4xl text-left mx-auto",
      "text-white/70",
      "text-left max-w-sm mx-0 md:text-sm my-2"
    )}>
      {children}
    </p>
  );
};

const SkeletonWithYoutube = ({
  setHoveredImage,
  hoveredImage,
  setIsYoutubeHovered,
  isYoutubeHovered,
}: {
  setHoveredImage: (image: string | null) => void;
  hoveredImage: string | null;
  setIsYoutubeHovered: (isHovered: boolean) => void;
  isYoutubeHovered: boolean;
}) => {
  const images = [
    "https://images.unsplash.com/photo-1517322048670-4fba75cbbb62?q=80&w=3000&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
    "https://images.unsplash.com/photo-1573790387438-4da905039392?q=80&w=3425&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
    "https://images.unsplash.com/photo-1555400038-63f5ba517a47?q=80&w=3540&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
    "https://images.unsplash.com/photo-1554931670-4ebfabf6e7a9?q=80&w=3387&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
    "https://images.unsplash.com/photo-1546484475-7f7bd55792da?q=80&w=2581&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
    "https://images.unsplash.com/photo-1573790387438-4da905039392?q=80&w=3425&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
    "https://images.unsplash.com/photo-1555400038-63f5ba517a47?q=80&w=3540&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
    "https://images.unsplash.com/photo-1554931670-4ebfabf6e7a9?q=80&w=3387&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D",
  ];

  const imageVariants = {
    initial: { scale: 1, rotate: 0, zIndex: 1 },
    hover: { scale: 1.1, rotate: 0, zIndex: 100 },
  };

  return (
    <div className="relative flex flex-row items-start p-8 gap-10 h-full overflow-hidden">
      <div className="flex-1">
        <div className="flex flex-wrap -ml-20 justify-center">
          {images.map((image, idx) => (
            <motion.div
              variants={imageVariants}
              initial="initial"
              whileHover="hover"
              key={idx}
              style={{
                rotate: Math.random() * 20 - 10,
              }}
              className="rounded-xl -mr-4 mt-4 p-1 bg-white/10 backdrop-blur-sm border border-white/10 flex-shrink-0 overflow-hidden cursor-pointer"
              onMouseEnter={() => setHoveredImage(image)}
              onMouseLeave={() => setHoveredImage(null)}
            >
              <Image
                src={image}
                alt={`image ${idx + 1}`}
                width="200"
                height="200"
                className="rounded-lg h-20 w-20 md:h-40 md:w-40 object-cover flex-shrink-0"
              />
            </motion.div>
          ))}
        </div>
      </div>

      <div className="flex-1 flex justify-center items-center">
        <motion.div
          initial={{ scale: 1 }}
          whileHover={{ scale: 1.05 }}
          onMouseEnter={() => setIsYoutubeHovered(true)}
          onMouseLeave={() => setIsYoutubeHovered(false)}
          className="w-full max-w-md h-full aspect-video bg-white/10 rounded-lg overflow-hidden cursor-pointer"
        >
          <iframe
            width="100%"
            height="100%"
            src="https://www.youtube.com/embed/dQw4w9WgXcQ"
            title="YouTube video player"
            frameBorder="0"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowFullScreen
          ></iframe>
        </motion.div>
      </div>
      {hoveredImage && (
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.8 }}
          className="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-50"
        >
          <Image
            src={hoveredImage}
            alt="Enlarged image"
            width="600"
            height="600"
            className="rounded-lg object-cover"
          />
        </motion.div>
      )}
      <div className="absolute left-0 z-[100] inset-y-0 w-20 bg-gradient-to-r from-black to-transparent h-full pointer-events-none" />
      <div className="absolute right-0 z-[100] inset-y-0 w-20 bg-gradient-to-l from-black to-transparent h-full pointer-events-none" />
    </div>
  );
};