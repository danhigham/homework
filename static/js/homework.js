$(document).ready(function() {
  $.ajax({
    type: "GET",
    contentType: "application/json",
    url: "./courses.json",
  }).done(function(courses) {
    i=0;
    var courseIDs = _.map(courses, function(c) {return c.id; }).join(",")

    loadSchedule(courseIDs);

    _.each(courses, function(course) {
      if (i == 0) course.active = true;
      courseDiv = renderTemplate("course_tmpl", course);
      $('#courses-carousel > .carousel-inner').append(courseDiv);

      $.ajax({
        type: "GET",
        contentType: "application/json",
        url: "./courses/" + course.id + "/assignments.json",
      }).done(function(assignments) {
        assignmentsUL = $('#course-' + course.id + '-assignments > ul');
        _.each(assignments, function(assignment) {
          var due = new Date(assignment.due_at)
          var now = new Date();
          if (due < now) return;
          li = renderTemplate("assignment_tmpl", assignment);
          assignmentsUL.append(li);
        });

        $("#course-" + course.id + "-assignments .due_at").each(function(d, i) { var b= new Date(i.innerText); $(i).text("Due: " + formatDate(b)); });
      });
      i++;
    })
  });

  $.ajax({
    type: "GET",
    contentType: "application/json",
    url: "./overdue.json",
  }).done(function(overdue) {
    _.each(overdue, function(assignment) {
      li = renderTemplate("overdue_tmpl", assignment);
      $('#overdue-items').append(li);
    });
  });

  function loadSchedule(courseIDs) {
    $.ajax({
      type: "GET",
      contentType: "application/json",
      url: "./courses/" + courseIDs + "/today.json",
    }).done(function(events) {
      var eventsUL = $('#schedule-items');
      _.each(events, function(event) {
        if (event.hidden) return;
        li = renderTemplate("event_tmpl", event);
        eventsUL.append(li);
      });
    });
  }

  function formatDate(date) {
    var monthNames = [
      "Jan", "Feb", "Mar",
      "Apr", "May", "Jun", "Jul",
      "Aug", "Sep", "Oct",
      "Nov", "Dec"
    ];

    var day = date.getDate();
    var monthIndex = date.getMonth();
    var year = date.getFullYear();

    return day + ' ' + monthNames[monthIndex] + ' ' + year;
  }

  function renderTemplate(name, ctx) {
    var template = $('#' + name).html();
    Mustache.parse(template);   // optional, speeds up future uses
    return Mustache.render(template, ctx);
  }
});
