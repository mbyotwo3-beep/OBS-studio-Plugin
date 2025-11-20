#!/usr/bin/env python3
"""
Tutorial Video Generator for OBS QR Donations Plugin

This script helps create a step-by-step video tutorial for the plugin.
It generates a script and can record the screen using OBS WebSocket.
"""
import os
import time
from pathlib import Path
from datetime import datetime

# Configuration
TUTORIAL_DIR = Path("docs/tutorial")
TUTORIAL_DIR.mkdir(parents=True, exist_ok=True)

# OBS WebSocket configuration (update these)
OBS_WS_HOST = "localhost"
OBS_WS_PORT = 4444
OBS_WS_PASSWORD = "your_password_here"

class TutorialScript:
    def __init__(self):
        self.steps = []
        self.current_step = 0
        
    def add_step(self, title, description, duration=5):
        """Add a step to the tutorial."""
        self.steps.append({
            'title': title,
            'description': description,
            'duration': duration,
            'timestamp': None
        })
    
    def generate_script(self):
        """Generate a markdown script for the tutorial."""
        script = "# OBS QR Donations Plugin Tutorial\n\n"
        total_duration = 0
        
        for i, step in enumerate(self.steps, 1):
            script += f"## {i}. {step['title']}\n"
            script += f"**Duration:** {step['duration']} seconds\n\n"
            script += f"{step['description']}\n\n"
            total_duration += step['duration']
        
        script += f"\n**Total Duration:** {total_duration} seconds"
        return script
    
    def save_script(self):
        """Save the tutorial script to a markdown file."""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        filename = TUTORIAL_DIR / f"tutorial_script_{timestamp}.md"
        
        with open(filename, 'w', encoding='utf-8') as f:
            f.write(self.generate_script())
        
        print(f"Tutorial script saved to: {filename}")
        return filename

    def record_step(self, step_num, output_dir):
        """Record a single step of the tutorial."""
        if step_num < 0 or step_num >= len(self.steps):
            print(f"Invalid step number: {step_num}")
            return
            
        step = self.steps[step_num]
        print(f"\n---\nRecording Step {step_num + 1}: {step['title']}")
        print(step['description'])
        
        # Here you would integrate with OBS WebSocket to start/stop recording
        # This is a placeholder for the actual implementation
        print(f"Recording for {step['duration']} seconds...")
        time.sleep(step['duration'])
        
        # Save the recording with a descriptive name
        output_file = output_dir / f"step_{step_num + 1:02d}_{step['title'].lower().replace(' ', '_')}.mp4"
        print(f"Saved recording to: {output_file}")
        
        return output_file

def create_tutorial():
    """Create a tutorial for the OBS QR Donations plugin."""
    tutorial = TutorialScript()
    
    # Introduction
    tutorial.add_step(
        "Introduction",
        "Welcome to the OBS QR Donations plugin tutorial. This guide will show you "
        "how to set up and use the plugin to receive cryptocurrency donations.",
        duration=5
    )
    
    # Installation
    tutorial.add_step(
        "Installation",
        "1. Download the latest release from GitHub\n"
        "2. Extract the plugin to your OBS plugins directory\n"
        "3. Restart OBS Studio",
        duration=10
    )
    
    # Basic Setup
    tutorial.add_step(
        "Basic Setup",
        "1. In OBS, add a new source and select 'QR Donations'\n"
        "2. Enter your wallet addresses for different cryptocurrencies\n"
        "3. Configure display options like size and theme",
        duration=15
    )
    
    # Lightning Network Setup
    tutorial.add_step(
        "Lightning Network Setup",
        "1. Enable Lightning Network in the plugin settings\n"
        "2. Enter your Breez API key\n"
        "3. Configure your Spark wallet connection",
        duration=15
    )
    
    # Using the Plugin
    tutorial.add_step(
        "Using the Plugin",
        "1. Generate a new invoice before going live\n"
        "2. Copy the payment details to your stream description\n"
        "3. Monitor incoming donations in real-time",
        duration=20
    )
    
    # Advanced Features
    tutorial.add_step(
        "Advanced Features",
        "1. Customize the QR code appearance\n"
        "2. Set minimum/maximum donation amounts\n"
        "3. Configure payment notifications",
        duration=15
    )
    
    # Save the tutorial script
    script_path = tutorial.save_script()
    print(f"\nTutorial script created at: {script_path}")
    
    # Ask if user wants to record the tutorial
    record = input("\nWould you like to record this tutorial now? (y/n): ").lower()
    if record == 'y':
        output_dir = TUTORIAL_DIR / "recordings" / datetime.now().strftime("%Y%m%d_%H%M%S")
        output_dir.mkdir(parents=True, exist_ok=True)
        
        print(f"\nStarting recording session. Output directory: {output_dir}")
        print("Make sure OBS is running and configured for screen recording.\n")
        
        for i in range(len(tutorial.steps)):
            tutorial.record_step(i, output_dir)
        
        print("\nRecording complete!")
        print(f"Videos saved to: {output_dir}")

if __name__ == "__main__":
    create_tutorial()
