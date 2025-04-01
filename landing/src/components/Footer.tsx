"use client";
import { motion } from "framer-motion";
import { Github, Twitter, Linkedin, Youtube } from "lucide-react";

export const Footer = () => {
  const socialLinks = [
    {
      icon: <Twitter className="size-5" />,
      href: "https://twitter.com/nixopus",
      label: "Twitter",
    },
    {
      icon: <Linkedin className="size-5" />,
      href: "https://linkedin.com/company/nixopus",
      label: "LinkedIn",
    },
    {
      icon: <Github className="size-5" />,
      href: "https://github.com/nixopus",
      label: "GitHub",
    },
    {
      icon: <Youtube className="size-5" />,
      href: "https://youtube.com/@nixopus",
      label: "YouTube",
    },
  ];

  return (
    <footer className="relative py-12">
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm"></div>
      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/10 to-transparent"></div>
      <div className="container relative">
        <div className="flex flex-col items-center gap-8 sm:flex-row sm:justify-between">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            viewport={{ once: true }}
            className="text-center sm:text-left"
          >
            <p className="text-white/60">
              © 2024 Nixopus. All rights reserved.
            </p>
          </motion.div>

          <motion.ul
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
            viewport={{ once: true }}
            className="flex items-center gap-4"
          >
            {socialLinks.map((link, index) => (
              <motion.li
                key={link.label}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.3 + index * 0.1 }}
                viewport={{ once: true }}
              >
                <a
                  href={link.href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="group p-2.5 rounded-lg bg-black/50 hover:bg-black/70 transition-all duration-300"
                  aria-label={link.label}
                >
                  <motion.div
                    whileHover={{ scale: 1.1 }}
                    whileTap={{ scale: 0.95 }}
                    className="text-white/80 group-hover:text-white transition-colors duration-300"
                  >
                    {link.icon}
                  </motion.div>
                </a>
              </motion.li>
            ))}
          </motion.ul>
        </div>
      </div>
    </footer>
  );
};
