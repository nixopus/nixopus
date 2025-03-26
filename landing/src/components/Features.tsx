"use client";
import React, { useState } from "react";
import { cn } from "@/lib/utils";
import Image from "next/image";
import { motion } from "framer-motion";

export function FeaturesSectionDemo() {
  const [hoveredImage, setHoveredImage] = useState<string | null>(null);
  const [isYoutubeHovered, setIsYoutubeHovered] = useState(false);

  const features = [
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
    <div className="relative z-20 py-10 lg:py-40 max-w-7xl mx-auto">
      <div className="px-8">
        <h4 className="text-3xl lg:text-5xl lg:leading-tight max-w-5xl mx-auto text-center tracking-tight font-medium text-black dark:text-white">
          Showcasing the coolness of Nixopus
        </h4>
        <p className="text-sm lg:text-base max-w-2xl my-4 mx-auto text-neutral-500 text-center font-normal dark:text-neutral-300">
          We are proud of what we have built, and we&apos;re always looking for ways to make the experience even better.
        </p>
      </div>
      <div className="relative">
        <div className="grid grid-cols-1 lg:grid-cols-1 mt-12 xl:border rounded-md dark:border-neutral-800">
          {features.map((feature) => (
            <FeatureCard key={feature.title} className={feature.className}>
              <FeatureTitle>{feature.title}</FeatureTitle>
              <FeatureDescription>{feature.description}</FeatureDescription>
              <div className="h-full w-full">{feature.skeleton}</div>
            </FeatureCard>
          ))}
        </div>
      </div>
    </div>
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
    <div className={cn(`p-4 sm:p-8 relative overflow-hidden`, className)}>
      {children}
    </div>
  );
};

const FeatureTitle = ({ children }: { children?: React.ReactNode }) => {
  return (
    <p className="max-w-5xl mx-auto text-left tracking-tight text-black dark:text-white text-xl md:text-2xl md:leading-snug">
      {children}
    </p>
  );
};

const FeatureDescription = ({ children }: { children?: React.ReactNode }) => {
  return (
    <p className={cn(
      "text-sm md:text-base max-w-4xl text-left mx-auto",
      "text-neutral-500 text-center font-normal dark:text-neutral-300",
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
              className="rounded-xl -mr-4 mt-4 p-1 bg-white dark:bg-neutral-800 dark:border-neutral-700 border border-neutral-100 flex-shrink-0 overflow-hidden cursor-pointer"
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
          className="w-full max-w-md  h-full aspect-video bg-gray-200 rounded-lg overflow-hidden cursor-pointer"
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
      <div className="absolute left-0 z-[100] inset-y-0 w-20 bg-gradient-to-r from-white dark:from-black to-transparent h-full pointer-events-none" />
      <div className="absolute right-0 z-[100] inset-y-0 w-20 bg-gradient-to-l from-white dark:from-black to-transparent h-full pointer-events-none" />
    </div>
  );
};