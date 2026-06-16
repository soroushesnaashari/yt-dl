### YouTube Video Downloader
Download any YouTube video in best quality directly into your GitHub repository.

<br>

### Usage
1. **Fork this repository**
2. Go to the **Actions** tab and enable workflows
3. Select **Download YouTube Video** (in the left sidebar)
4. Then **Run workflow** (in the right part of the blue section)
5. Paste the YouTube URL and run
6. After completion, the `download` folder will contain video chunks (e.g., `video.zip.00`, `video.zip.01`)

<br>

### Reassemble the video
**macOS / Linux** (Terminal):  
`cat video.zip.0* > video.zip`  
Then unzip `video.zip`.

**Windows** (Command Prompt):  
`copy /b download\video.zip.00 + download\video.zip.01 video.zip`  
Then extract `video.zip`.

<br>

---
<sub>*This project is for educational purposes only, please respect YouTube's Terms of Service.<br>
No local tools required, everything runs on GitHub Actions.*<sub>
