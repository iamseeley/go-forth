async function fetchLastUpdateTime(repoOwner, repoName) {
    const cacheKey = `lastUpdated-${repoOwner}-${repoName}`;
    const cachedData = localStorage.getItem(cacheKey);
    const cacheDuration = 60 * 60 * 1000; // Cache duration in milliseconds, e.g., 1 hour

    if (cachedData) {
        const { lastUpdated, timestamp } = JSON.parse(cachedData);
        const isCacheValid = (new Date() - new Date(timestamp)) < cacheDuration;

        if (isCacheValid) {
            return lastUpdated;
        }
    }

    try {
        const response = await fetch(`https://api.github.com/repos/${repoOwner}/${repoName}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const repoData = await response.json();
        const date = new Date(repoData.pushed_at);
        const formattedDate = date.toLocaleDateString('en-US', { 
            month: 'long', 
            day: 'numeric',
        }) + ' at ' + date.toLocaleTimeString('en-US', {
            hour: 'numeric',
            minute: '2-digit',
            hour12: true
        });

        localStorage.setItem(cacheKey, JSON.stringify({ lastUpdated: formattedDate, timestamp: new Date() }));
        return formattedDate;
    } catch (error) {
        console.error('Error fetching repository data:', error);
        return 'Unable to fetch update time.';
    }
}

document.addEventListener('DOMContentLoaded', async (event) => {
    const lastUpdated = await fetchLastUpdateTime('iamseeley', 'tseeley');
    document.getElementById('last-updated').textContent = `Last Updated: ${lastUpdated}`;
});
