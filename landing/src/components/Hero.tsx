"use client";
import { motion } from "framer-motion";
import { ArrowRightIcon, ChevronRight } from "lucide-react";
import Link from "next/link";
import AnimatedShinyText from "./shinyText";
import { cn } from "./lib/utils";
import { BackgroundLines } from "./ui/background-lines";

export default function Hero() {
  return (
    <section >
      <div className="relative h-full py-10 bg-gradient-to-b from-gray-900 via-gray-950 to-black">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_120%,rgba(120,119,198,0.1),rgba(255,255,255,0))]"></div>
        <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
        <div className="container z-10 flex flex-col">
          <div className=" grid grid-cols-1">
            <div className="flex flex-col items-center gap-6 pb-8 text-center">
              <Link href="/components/text">
                <div
                  className={cn(
                    "group rounded-full border border-white/10 bg-gradient-to-r from-indigo-500/10 to-purple-500/10 text-base text-white transition-all ease-in hover:cursor-pointer hover:from-indigo-500/20 hover:to-purple-500/20",
                  )}
                >
                  <AnimatedShinyText className="inline-flex items-center justify-center px-4 py-1 transition ease-out hover:text-indigo-300 hover:duration-300">
                    <span>✨ Version alpha is here </span>
                    <ArrowRightIcon className="ml-1 size-3 transition-transform duration-300 ease-in-out group-hover:translate-x-0.5" />
                  </AnimatedShinyText>
                </div>
              </Link>
              <BackgroundLines className="relative flex flex-col gap-4 md:items-center lg:flex-row">
                <h1 className="relative mx-0 max-w-[43.5rem] text-balance pt-5 text-left text-5xl font-extrabold tracking-tight text-white dark:text-white sm:text-7xl md:mx-auto md:px-4 md:py-2 md:text-center md:text-7xl lg:text-7xl">
                  <span className="bg-gradient bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 mr-2">Nixopus</span>
                  <br />
                  All in one tool for managing your vps
                </h1>
                <span className="text-neutral-90 absolute -top-3.5 left-0 z-10 rotate-3 whitespace-nowrap rounded-full bg-gradient-to-r from-indigo-500 to-purple-500 px-2.5 py-1 text-[11px] font-semibold uppercase leading-5 tracking-wide text-white md:top-12 md:-rotate-12">
                  Self hostable
                </span>
              </BackgroundLines>
              <p className="max-w-[64rem] text-balance text-lg tracking-tight text-gray-300 dark:text-gray-300 md:text-xl">
                Nixopus boosts your productivity for managing vps graphically with built in features such as{" "}
                <strong className="text-indigo-300">Self hosting capabilities</strong>, <strong className="text-purple-300">Open source Marketplace</strong>,{" "}
                <strong className="text-pink-300">File Manager</strong>, and{" "}
                <strong className="text-indigo-300">Mail Server </strong>.
                <br />
                <span className="text-gray-400">100% open-source, with love.</span>
              </p>

              <div className="flex flex-col gap-4 lg:flex-row" >
                <div className="flex flex-col gap-4 md:flex-row">
                  <Link
                    href="/components"
                    className={cn(
                      "group relative inline-flex items-center rounded-full border bg-gradient-to-r from-indigo-500 to-purple-500 border-transparent px-4 py-2 text-sm font-semibold tracking-tighter text-white transition-all ease-in hover:from-indigo-600 hover:to-purple-600 hover:text-white focus-visible:outline-none focus-visible:ring focus-visible:ring-indigo-500 focus-visible:ring-offset-2 focus-visible:ring-offset-indigo-100",
                    )}
                  >
                    Self host
                    <ChevronRight className="ml-1 size-4 flex-shrink-0 transition-all duration-300 ease-out group-hover:translate-x-1" />
                  </Link>
                  <Link
                    href="/docs/installation"
                    className={cn(
                      "group relative inline-flex items-center rounded-full border border-white/20 px-4 py-2 text-sm font-semibold tracking-tighter text-white transition-all ease-in hover:bg-gradient-to-r hover:from-indigo-500/20 hover:to-purple-500/20 hover:border-transparent focus-visible:outline-none focus-visible:ring focus-visible:ring-indigo-500 focus-visible:ring-offset-2 focus-visible:ring-offset-indigo-100",
                    )}
                  >
                    Try cloud version
                    <ChevronRight className="ml-1 size-4 flex-shrink-0 transition-all duration-300 ease-out group-hover:translate-x-1" />
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div className="container relative mx-auto mt-32 w-full max-w-[1300px]">
          <motion.span
            animate={["initial"]}
            whileHover={["hover"]}
            variants={{
              hover: {
                scale: 1.1,
                rotate: -6,
                transition: {
                  duration: 0.2,
                },
              },
              initial: {
                y: [-8, 8],
                transition: {
                  duration: 2,
                  repeat: Infinity,
                  repeatType: "reverse",
                },
              },
            }}
            className="absolute -top-16 left-0 right-auto cursor-pointer lg:-top-20"
          >
            <span className="flex items-center">
              <span className="mt-3 inline-block whitespace-nowrap rounded-full bg-gradient-to-r from-indigo-500/20 to-purple-500/20 px-2.5 py-1 text-[11px] font-semibold uppercase leading-5 tracking-wide text-white">
                Explore Features
              </span>
              <svg
                className="mr-6 h-8 w-14 [transform:rotateY(180deg)rotateX(0deg)]"
                width="45"
                height="25"
                viewBox="0 0 45 25"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M43.2951 3.47877C43.8357 3.59191 44.3656 3.24541 44.4788 2.70484C44.5919 2.16427 44.2454 1.63433 43.7049 1.52119L43.2951 3.47877ZM4.63031 24.4936C4.90293 24.9739 5.51329 25.1423 5.99361 24.8697L13.8208 20.4272C14.3011 20.1546 14.4695 19.5443 14.1969 19.0639C13.9242 18.5836 13.3139 18.4152 12.8336 18.6879L5.87608 22.6367L1.92723 15.6792C1.65462 15.1989 1.04426 15.0305 0.563943 15.3031C0.0836291 15.5757 -0.0847477 16.1861 0.187863 16.6664L4.63031 24.4936ZM43.7049 1.52119C32.7389 -0.77401 23.9595 0.99522 17.3905 5.28788C10.8356 9.57127 6.58742 16.2977 4.53601 23.7341L6.46399 24.2659C8.41258 17.2023 12.4144 10.9287 18.4845 6.96211C24.5405 3.00476 32.7611 1.27399 43.2951 3.47877L43.7049 1.52119Z"
                  fill="currentColor"
                  className="fill-gray-300 dark:fill-gray-700"
                />
              </svg>
            </span>
          </motion.span>
        </div>
      </div>
    </section>
  );
}