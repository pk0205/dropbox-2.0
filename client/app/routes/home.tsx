import type { Route } from "./+types/home";
import { Button } from "~/components/ui/button";
import { ArrowRight } from "lucide-react";
import { useNavigate } from "react-router";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "New React Router App" },
    { name: "description", content: "Welcome to React Router!" },
  ];
}

export default function Home() {
  const navigate = useNavigate();
  return (
    <div className="flex min-h-svh flex-col p-8 items-center">
      <h1 className="text-5xl font-bold">Welcome to Dropbox 2.0</h1>
      <h1 className="text-3xl font-semibold">
        Storing everything for you and your business needs. All in one place.
      </h1>
      <p className="pb-16">
        Enhance your personal storage with Dropbox, offering a simple and
        efficient way to upload, organize, and access files from anywhere.
        Securely store important documents and media, and experience the
        convenience of easy file management and sharing in one centralized
        solution.
      </p>

      <Button
        onClick={() => navigate("/dashboard")}
        className="flex cursor-pointer bg-blue-500 p-5 w-fit rounded-xl hover:tracking-widest transition-all duration-300 hover:font-bold"
      >
        Try it for free! <ArrowRight className="ml-6" />
      </Button>
    </div>
  );
}
