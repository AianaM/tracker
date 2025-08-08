(() => {
    const path = (() => {
        const re = RegExp(/\/worklog\/(?<createdBy>[^\/]+)\/((from\/(?<from>[^\/]+)\/to\/(?<to>[^\/]+))|(?<preset>[^\/]+))(\/show\/((from\/(?<showFrom>[^\/]+)\/to\/(?<showTo>[^\/]+))|(?<showPreset>[^\/]+)))?/);
        const params = re.exec(window.location.pathname);

        const getCreatedByPath = (createdBy) => `/worklog/${createdBy}`;
        const getWorklogPath = (period) => `/${period.preset || period.from + "/" + period.to}`;
        const getShowPath = (period) => {
            const params = period.showPreset ? period.preset : period.from ? period.from + "/" + period.to : "";
            return params ? `/show/${params}` : "";
        };
        const createdByPath = getCreatedByPath(params);
        const worklogPath = getWorklogPath(params);
        const showPath = getShowPath({ preset: params.showPreset, from: params.showFrom, to: params.showTo });
        return {
            getCreatedByPath,
            getWorklogPath,
            getShowPath,
            links: {
                getWorklogLink: (period) => createdByPath + getWorklogPath(period) + showPath,
                getShowLink: (period) => createdByPath + worklogPath + getShowPath(period),
            }
        };
    })();

    const appendHeaderLinks = () => {
        const periods = [{ preset: "today", value: "Сегодня" }, { preset: "currentWeek", value: "Эта неделя" }, { preset: "currentMonth", value: "Этот месяц" }];
        const worklog = (periods) => {
            const el = document.querySelector(".header .worklog");
            periods.forEach((period) => {
                const link = document.createElement("a");
                link.href = path.links.getWorklogLink(period);
                link.textContent = period.value;
                el.appendChild(link);
            });
        };
        const show = (periods) => {
            const el = document.querySelector(".header .show");
            periods.forEach((period) => {
                const link = document.createElement("a");
                link.href = path.links.getShowLink(period);
                link.textContent = period.value;
                el.appendChild(link);
            });
        };
        worklog(periods);
        show(periods);
    };
    appendHeaderLinks();

    const setInputsValues = () => {
        const dateToISOString = (date) => {
            return date.toISOString().substring(0, 10);
        }
        const setWorklogInputValues = () => {
            document.getElementById("worklog-start").value = dateToISOString(query.createdAt.from);
            document.getElementById("worklog-end").value = dateToISOString(query.createdAt.to);
        }
        const setShowInputValues = () => {
            document.getElementById("show-start").value = dateToISOString(query.show.from);
            document.getElementById("show-end").value = dateToISOString(query.show.to);
        }
        setWorklogInputValues();
        setShowInputValues();
    }
    setInputsValues();
})()

const onShowSubmit = () => {
    const start = document.getElementById("show-start").value;
    const end = document.getElementById("show-end").value;
    if (start && end) {
        window.location.href = path.links.getShowLink({ from: start, to: end });
    } else {
        alert("Please fill in both start and end dates.");
    }
}
function onWorklogSubmit() {
    const start = document.getElementById("worklog-start").value;
    const end = document.getElementById("worklog-end").value;
    if (start && end) {
        window.location.href = path.links.getWorklogLink({ from: start, to: end });
    } else {
        alert("Please fill in both start and end dates.");
    }
}