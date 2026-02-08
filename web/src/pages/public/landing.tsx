import { LandingNav } from "@/components/landing/landing-nav";
import { Hero } from "@/components/landing/hero";
import { FeaturesGrid } from "@/components/landing/features-grid";
import { PricingTable } from "@/components/landing/pricing-table";
import { Footer } from "@/components/landing/footer";

export function Component() {
  return (
    <>
      <LandingNav />
      <Hero />
      <FeaturesGrid />
      <PricingTable />
      <Footer />
    </>
  );
}
