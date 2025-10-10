document.addEventListener('DOMContentLoaded', () => {
    const searchBtn = document.getElementById('search-btn');
    const addressSearch = document.getElementById('address-search');
    const searchResults = document.getElementById('search-results');
    const generateIcsBtn = document.getElementById('generate-ics-btn');
    const uprnIcsInput = document.getElementById('uprn-ics-input');
    const icsLinkContainer = document.getElementById('ics-link-container');

    const generateIcsLink = () => {
        const uprn = uprnIcsInput.value.trim();
        const sbcBaseUrl = 'https://www.swindon.gov.uk/info/20122/rubbish_and_recycling_collection_days';

        if (uprn) {
            const icsLink = `${window.location.origin}/${uprn}/ics`;
            const sbcCheckLink = `${sbcBaseUrl}?uprn=${uprn}&addressList=${uprn}&uprnSubmit=Yes`;
            icsLinkContainer.innerHTML = `<p><strong>Your calendar link is ready!</strong></p>
                                          <p><a href="${icsLink}" target="_blank">${icsLink}</a></p>
                                          <p>You can also <a href="${sbcCheckLink}" target="_blank">check your collection days on the SBC website</a> to verify the data.</p>
                                          <p>Now, follow the instructions in Step 3 below to add it to your calendar.</p>`;
        } else {
            icsLinkContainer.innerHTML = '';
        }
    };

    uprnIcsInput.addEventListener('input', generateIcsLink);

    const performSearch = () => {
        const query = addressSearch.value.trim();
        if (query) {
            searchResults.textContent = 'Searching...';
            fetch(`/search-address?q=${encodeURIComponent(query)}`)
                .then(response => response.json())
                .then(data => {
                    searchResults.innerHTML = '';
                    if (data && data.length > 0) {
                        const ul = document.createElement('ul');
                        data.forEach(item => {
                            const li = document.createElement('li');
                            li.innerHTML = `<strong>${item.address}</strong> (UPRN: ${item.uprn})`;
                            li.dataset.uprn = item.uprn;
                            li.addEventListener('click', () => {
                                uprnIcsInput.value = item.uprn;
                                generateIcsLink();
                                // Optional: scroll to the link container
                                document.getElementById('ics-generator').scrollIntoView({ behavior: 'smooth' });
                            });
                            ul.appendChild(li);
                        });
                        searchResults.appendChild(ul);
                    } else {
                        searchResults.textContent = 'No results found.';
                    }
                })
                .catch(error => {
                    console.error('Error fetching search results:', error);
                    searchResults.textContent = 'Failed to fetch results.';
                });
        }
    };

    searchBtn.addEventListener('click', performSearch);

    addressSearch.addEventListener('keypress', (event) => {
        if (event.key === 'Enter') {
            event.preventDefault();
            performSearch();
        }
    });

    generateIcsBtn.addEventListener('click', generateIcsLink);
});