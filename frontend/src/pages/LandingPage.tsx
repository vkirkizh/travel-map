import {useEffect} from "react";
import "./LandingPage.css";

export function LandingPage() {
  useEffect(() => {
    document.title = "Travel Map — Share your travels";
  }, []);

  return (
    <div className="landing-page">
      <div className="landing-card">
        <div className="landing-eyebrow">Travel Map</div>
        <h1>Share the places you have visited.</h1>
        <p>
          A personal travel map with visited cities, landmarks and travel statistics.
        </p>
        <p>
          Created by Valery Kirkizh: <a href="mailto:valery@kirkizh.com">Email</a> &bull;&nbsp;<a href="https://www.linkedin.com/in/vkirkizh/" rel="me">LinkedIn</a> &bull;&nbsp;<a href="https://github.com/vkirkizh" rel="me">GitHub</a>
        </p>
        <div className="landing-actions">
          <a href="/valery/" className="landing-link">
            View demo map
          </a>
          <a href="/app/" className="landing-link landing-link-secondary">
            Log In / Sign Up
          </a>
        </div>
      </div>
    </div>
  );
}
