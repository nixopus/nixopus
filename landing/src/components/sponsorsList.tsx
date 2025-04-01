import Image from 'next/image'
import acmeLogo from "../assets/images/acme.png";
import quantumLogo from "../assets/images/quantum.png";
import echoLogo from "../assets/images/echo.png";
import celestialLogo from "../assets/images/celestial.png";
import pulseLogo from "../assets/images/pulse.png";
import apexLogo from "../assets/images/apex.png";

export default function LogoCarousel() {
  const logos = [
    { src: acmeLogo, alt: "Acme Logo" },
    { src: quantumLogo, alt: "Quantum Logo" },
    { src: echoLogo, alt: "Echo Logo" },
    { src: celestialLogo, alt: "Celestial Logo" },
    { src: pulseLogo, alt: "Pulse Logo" },
    { src: apexLogo, alt: "Apex Logo" },
  ]

  return (
    <div className="relative py-12 overflow-hidden bg-gradient-to-r from-indigo-500/5 to-purple-500/5 backdrop-blur-sm">
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#4f4f4f2e_1px,transparent_1px),linear-gradient(to_bottom,#4f4f4f2e_1px,transparent_1px)] bg-[size:14px_24px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)]"></div>
      
      <div className="relative z-10">
        <div className="w-full inline-flex flex-nowrap overflow-hidden [mask-image:_linear-gradient(to_right,transparent_0,_black_128px,_black_calc(100%-128px),transparent_100%)]">
          <ul className="flex items-center justify-center md:justify-start [&_li]:mx-8 [&_img]:max-w-none animate-infinite-scroll">
            {logos.map((logo, index) => (
              <li key={index} className="group">
                <div className="relative p-4 transition-all duration-300 group-hover:scale-105">
                  <div className="absolute inset-0 bg-gradient-to-r from-indigo-500/10 to-purple-500/10 rounded-lg blur-xl group-hover:opacity-100 opacity-0 transition-opacity duration-300"></div>
                  <Image 
                    src={logo.src} 
                    alt={logo.alt} 
                    className="relative z-10 filter grayscale hover:grayscale-0 transition-all duration-300"
                  />
                </div>
              </li>
            ))}
          </ul>
          <ul className="flex items-center justify-center md:justify-start [&_li]:mx-8 [&_img]:max-w-none animate-infinite-scroll" aria-hidden="true">
            {logos.map((logo, index) => (
              <li key={index} className="group">
                <div className="relative p-4 transition-all duration-300 group-hover:scale-105">
                  <div className="absolute inset-0 bg-gradient-to-r from-indigo-500/10 to-purple-500/10 rounded-lg blur-xl group-hover:opacity-100 opacity-0 transition-opacity duration-300"></div>
                  <Image 
                    src={logo.src} 
                    alt={logo.alt} 
                    className="relative z-10 filter grayscale hover:grayscale-0 transition-all duration-300"
                  />
                </div>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </div>
  )
}