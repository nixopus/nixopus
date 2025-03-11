import React from 'react';
import { motion } from 'framer-motion';

const WireEffect = ({ children }: { children: React.ReactNode }) => {
    const lineVariants = {
        hidden: { pathLength: 0, opacity: 0 },
        visible: {
            pathLength: 1,
            opacity: 1,
            transition: {
                duration: 2,
                ease: "easeInOut",
                repeat: Infinity,
                repeatType: "reverse"
            }
        }
    };

    return (
        <div className="relative">
            {children}
            <svg className="absolute inset-0 w-full h-full pointer-events-none" xmlns="http://www.w3.org/2000/svg">
                <motion.rect
                    x="0"
                    y="0"
                    width="100%"
                    height="100%"
                    fill="none"
                    stroke="#00FFFF"
                    strokeWidth="2"
                    initial="hidden"
                    animate="visible"
                    variants={lineVariants}
                />
            </svg>
        </div>
    );
};

export default WireEffect;