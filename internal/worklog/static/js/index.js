(() => {
    const pathname = {
        worklog: RegExp(/\/worklog((\/from\/[\d-]+\/to\/[\d-]+)|(\/\w+))/).exec(window.location.pathname)?.at(0),
        show: RegExp(/\/show((\/from\/[\d-]+\/to\/[\d-]+)|(\/\w+))/).exec(window.location.pathname)?.at(0)
    }
    const appendWorklogLinks = (periods) => {
        const el = document.querySelector(".header .worklog");
        periods.forEach((period) => {
            const link = document.createElement("a");
            link.href = getWorklogLink(period.key);
            link.textContent = period.value;
            el.appendChild(link);
        });
    }
    const appendShowLinks = (periods) => {
        const el = document.querySelector(".header .show");
        periods.forEach((period) => {
            const link = document.createElement("a");
            link.href = getShowLink(period.key);
            link.textContent = period.value;
            el.appendChild(link);
        });
    }
    const getWorklogLink = (period) => `worklog/${period}${pathname.show ?? ""}`;
    const getShowLink = (period) => `${pathname.worklog}/show/${period}`;
    const periods = [{ key: "today", value: "Сегодня" }, { key: "currentWeek", value: "Эта неделя" }, { key: "currentMonth", value: "Этот месяц" }];
    appendWorklogLinks(periods);
    appendShowLinks(periods);

    const dateToISOString = (date) => {
        return date.toISOString().substring(0, 10);
    }

    const getTimespan = (req) => {
        const date = window.location.pathname.match(req);
        if (!date || date.length < 3) {
            const today = new Date();
            const end = new Date(today.getTime() + (7 * 24 * 60 * 60 * 1000)); // Add one week to today
            return {
                start: dateToISOString(today),
                end: dateToISOString(end)
            };
        }
        const start = date[1];
        const end = date[2];
        return {
            start,
            end
        };
    }
    const setWorklogInputValues = () => {
        const timespan = getTimespan(/\/worklog\/from\/([\d-]+)\/to\/([\d-]+)/);
        document.getElementById("worklog-start").value = timespan.start;
        document.getElementById("worklog-end").value = timespan.end;
    }
    const setShowInputValues = () => {
        const timespan = getTimespan(/\/show\/from\/([\d-]+)\/to\/([\d-]+)/);
        document.getElementById("show-start").value = timespan.start;
        document.getElementById("show-end").value = timespan.end;
    }
    setWorklogInputValues();
    setShowInputValues();
})()

const onShowSubmit = () => {
    const start = document.getElementById("show-start").value;
    const end = document.getElementById("show-end").value;
    if (start && end) {
        window.location.href = getShowLink(`from/${start}/to/${end}`);
    } else {
        alert("Please fill in both start and end dates.");
    }
}
function onWorklogSubmit() {
    const start = document.getElementById("worklog-start").value;
    const end = document.getElementById("worklog-end").value;
    if (start && end) {
        window.location.href = getWorklogLink(`from/${start}/to/${end}`);
    } else {
        alert("Please fill in both start and end dates.");
    }
}