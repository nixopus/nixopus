import { Banner } from "@/components/Banner";
import { Navbar } from "@/components/Navbar";
import { LogoTicker } from "@/components/LogoTicker";
import { ProductShowcase } from "@/components/ProductShowcase";
import { FAQs } from "@/components/FAQs";
import { CallToAction } from "@/components/CallToAction";
import { Footer } from "@/components/Footer";
import Hero from "@/components/Hero";
import Terminal from "@/components/Terminal";
import SelfHost from "@/components/SelfHost";
import MarketPlace from "@/components/MarketPlace";
import CronJob from "@/components/CronJob";
import { CardWithEffect } from "@/components/bentogrid";
import Firewall from "@/components/Firewall";
import { FeaturesSectionDemo } from "@/components/Features";

export default function Home() {
  return (
    <>
      <div className="overflow-x-hidden bg-black">
        <Banner />
        <Navbar />
        <Hero />
        <LogoTicker />
        <div className="bg-black container p-20">
          <h2 className="text-center font-bold text-5xl sm:text-6xl tracking-tighter text-white">Packed with features</h2>
          <div className='max-w-xl mx-auto'>
            <p className="text-center mt-5 text-xl text-white/70 mb-10">
              You expect cool features and nixopus is here to fulfill that expectation.
            </p>
          </div>
          <div className="grid grid-cols-2 gap-10 mb-10">
            <CardWithEffect>
              <Terminal />
            </CardWithEffect>
            <CardWithEffect>
              <SelfHost />
            </CardWithEffect>
          <CardWithEffect>
            <Firewall />
          </CardWithEffect>
          <CardWithEffect>
            <Firewall />
          </CardWithEffect>
          </div>
          <div className="grid grid-cols-2 gap-10 mb-10 mt-10">
            <CardWithEffect>
              <MarketPlace />
            </CardWithEffect>
            <CardWithEffect>
              <CronJob />
            </CardWithEffect>
          </div>
        </div>

        <FeaturesSectionDemo />

        <ProductShowcase />
        <FAQs />
        <CallToAction />
      </div>
      <Footer />
    </>
  );
}
