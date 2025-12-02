package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	application "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Application"
	chat "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Chat"
	course "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Course"
	department "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Department"
	dispute "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Dispute"
	invitation "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Invitation"
	milestone "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Milestone"
	notification "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Notification"
	organization "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Organization"
	portfolio "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Portfolio"
	project "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Project"
	student "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Student"
	supervisor "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/Supervisor"
	supervisorrequest "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/SupervisorRequest"
	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Seed Organizations
func seedOrganizations(db *gorm.DB, count int, passwordHash string) []organization.Organization {
	var orgs []organization.Organization

	for i := 0; i < count; i++ {
		var orgName string
		var orgType string

		if i < len(ugandanUniversities) {
			orgName = ugandanUniversities[i]
			orgType = "university"
		} else if i-len(ugandanUniversities) < len(ugandanCompanies) {
			orgName = ugandanCompanies[i-len(ugandanUniversities)]
			orgType = "partner"
		} else {
			orgName = fmt.Sprintf("Organization %d", i+1)
			orgType = "university"
			if i%2 == 1 {
				orgType = "partner"
			}
		}

		// Create admin user for organization
		adminUser := user.User{
			Role:     "university-admin",
			Email:    fmt.Sprintf("admin@%s.com", sanitizeEmail(orgName)),
			Name:     fmt.Sprintf("%s %s", getRandomFirstName(), getRandomLastName()),
			Password: passwordHash,
			Profile: user.Profile{
				Phone:    fmt.Sprintf("+2567%08d", rand.Intn(100000000)),
				Location: "Kampala, Uganda",
			},
		}
		if orgType == "partner" {
			adminUser.Role = "partner"
		}
		db.Create(&adminUser)

		org := organization.Organization{
			Name:       orgName,
			Type:       orgType,
			IsApproved: true,
			UserID:     adminUser.ID,
			Website:    fmt.Sprintf("https://www.%s.com", sanitizeEmail(orgName)),
			Address:    fmt.Sprintf("%d %s Street, Kampala, Uganda", rand.Intn(999)+1, getRandomStreetName()),
		}
		db.Create(&org)
		orgs = append(orgs, org)
	}

	return orgs
}

// Seed Departments
func seedDepartments(db *gorm.DB, orgs []organization.Organization) []department.Department {
	var depts []department.Department

	// Only create departments for universities
	universityOrgs := []organization.Organization{}
	for _, org := range orgs {
		if org.Type == "university" {
			universityOrgs = append(universityOrgs, org)
		}
	}

	for _, org := range universityOrgs {
		// Create 3-8 departments per university
		numDepts := rand.Intn(6) + 3
		for i := 0; i < numDepts && i < len(ugandanDepartments); i++ {
			dept := department.Department{
				Name:           ugandanDepartments[(len(depts)+i)%len(ugandanDepartments)],
				OrganizationID: uint(org.ID),
			}
			db.Create(&dept)
			depts = append(depts, dept)
		}
	}

	return depts
}

// Seed Courses
func seedCourses(db *gorm.DB, depts []department.Department) []course.Course {
	var courses []course.Course

	for _, dept := range depts {
		// Create 2-5 courses per department
		numCourses := rand.Intn(4) + 2
		for i := 0; i < numCourses && i < len(ugandanCourses); i++ {
			crs := course.Course{
				Name:         ugandanCourses[(len(courses)+i)%len(ugandanCourses)],
				DepartmentID: uint(dept.ID),
			}
			db.Create(&crs)
			courses = append(courses, crs)
		}
	}

	return courses
}

// Seed Users
func seedUsers(db *gorm.DB, orgs []organization.Organization, courses []course.Course, passwordHash string, count int) []user.User {
	var users []user.User
	roles := []string{"student", "supervisor", "partner"}

	for i := 0; i < count; i++ {
		role := roles[rand.Intn(len(roles))]
		var courseID uint

		// Assign course to students
		if role == "student" && len(courses) > 0 {
			courseID = uint(courses[rand.Intn(len(courses))].ID)
		}

		firstName := getRandomFirstName()
		lastName := getRandomLastName()
		email := fmt.Sprintf("%s.%s.%d@example.com", sanitizeEmail(firstName), sanitizeEmail(lastName), i+1)

		userSkills := []string{}
		numSkills := rand.Intn(5) + 3
		for j := 0; j < numSkills; j++ {
			userSkills = append(userSkills, skills[rand.Intn(len(skills))])
		}
		skillsJSON, _ := json.Marshal(userSkills)

		usr := user.User{
			Role:     role,
			Email:    email,
			Name:     fmt.Sprintf("%s %s", firstName, lastName),
			Password: passwordHash,
			Profile: user.Profile{
				Phone:    fmt.Sprintf("+2567%08d", rand.Intn(100000000)),
				Location: "Kampala, Uganda",
				Bio:      fmt.Sprintf("Experienced %s with expertise in various technologies", role),
				Skills:   datatypes.JSON(skillsJSON),
			},
			CourseID: courseID,
		}
		db.Create(&usr)
		users = append(users, usr)
	}

	return users
}

// Seed Students
func seedStudents(db *gorm.DB, users []user.User, courses []course.Course) []student.Student {
	var students []student.Student

	for _, usr := range users {
		if usr.Role == "student" && usr.CourseID > 0 {
			std := student.Student{
				UserID:   usr.ID,
				CourseID: uint(usr.CourseID),
			}
			db.Create(&std)
			students = append(students, std)
		}
	}

	return students
}

// Seed Supervisors
func seedSupervisors(db *gorm.DB, users []user.User, depts []department.Department) []supervisor.Supervisor {
	var supervisors []supervisor.Supervisor

	// Get departments
	deptMap := make(map[uint]department.Department)
	for _, dept := range depts {
		deptMap[dept.ID] = dept
	}

	for _, usr := range users {
		if usr.Role == "supervisor" && usr.OrgID != nil {
			// Find a department from the same organization
			var matchingDept *department.Department
			for _, dept := range depts {
				if dept.OrganizationID == *usr.OrgID {
					matchingDept = &dept
					break
				}
			}
			if matchingDept != nil {
				sup := supervisor.Supervisor{
					UserID:       usr.ID,
					DepartmentID: matchingDept.ID,
				}
				db.Create(&sup)
				supervisors = append(supervisors, sup)
			}
		}
	}

	return supervisors
}

// Seed Projects
func seedProjects(db *gorm.DB, users []user.User, depts []department.Department, supervisors []supervisor.Supervisor, count int) []project.Project {
	var projects []project.Project

	// Get partner users
	partnerUsers := []user.User{}
	for _, usr := range users {
		if usr.Role == "partner" {
			partnerUsers = append(partnerUsers, usr)
		}
	}

	statuses := []string{"pending", "in-progress", "completed", "on-hold"}
	currencies := []string{"USD", "UGX"}

	for i := 0; i < count; i++ {
		if len(partnerUsers) == 0 || len(depts) == 0 {
			continue
		}

		partner := partnerUsers[rand.Intn(len(partnerUsers))]
		dept := depts[rand.Intn(len(depts))]

		title := projectTitles[i%len(projectTitles)]
		description := projectDescriptions[i%len(projectDescriptions)]

		// Random skills
		projectSkills := []string{}
		numSkills := rand.Intn(5) + 3
		for j := 0; j < numSkills; j++ {
			projectSkills = append(projectSkills, skills[rand.Intn(len(skills))])
		}
		skillsJSON, _ := json.Marshal(projectSkills)

		// Random budget
		currency := currencies[rand.Intn(len(currencies))]
		budgetValue := uint(rand.Intn(50000) + 1000)
		if currency == "UGX" {
			budgetValue = uint(rand.Intn(200000000) + 1000000)
		}

		// Random deadline (30-180 days from now)
		deadline := time.Now().AddDate(0, 0, rand.Intn(150)+30).Format("2006-01-02")

		// Optional supervisor
		var supervisorID *uint
		if len(supervisors) > 0 && rand.Float32() < 0.3 {
			sup := supervisors[rand.Intn(len(supervisors))]
			if sup.DepartmentID == dept.ID {
				supervisorID = &sup.UserID
			}
		}

		proj := project.Project{
			DepartmentID: int(dept.ID),
			Title:        title,
			Description:  description,
			Skills:       datatypes.JSON(skillsJSON),
			Budget: project.Budget{
				Currency: currency,
				Value:    budgetValue,
			},
			Deadline:     deadline,
			Capacity:     uint(rand.Intn(5) + 1),
			Status:       statuses[rand.Intn(len(statuses))],
			UserID:       partner.ID,
			SupervisorID: supervisorID,
		}
		db.Create(&proj)
		projects = append(projects, proj)
	}

	return projects
}

// Seed Applications
func seedApplications(db *gorm.DB, users []user.User, projects []project.Project, count int) []application.Application {
	var applications []application.Application

	// Get student users
	studentUsers := []user.User{}
	for _, usr := range users {
		if usr.Role == "student" {
			studentUsers = append(studentUsers, usr)
		}
	}

	// Define status distribution following the workflow:
	// SUBMITTED (40%) → SHORTLISTED (20%) → OFFERED (15%) → ASSIGNED (10%) / DECLINED (5%)
	// Also include: WAITLIST (5%), REJECTED (5%)
	statusWeights := map[string]int{
		"SUBMITTED":   40,
		"SHORTLISTED": 20,
		"OFFERED":     15,
		"ASSIGNED":    10,
		"DECLINED":    5,
		"WAITLIST":    5,
		"REJECTED":    5,
	}

	// Helper function to select status based on weights
	getWeightedStatus := func() string {
		total := 0
		for _, weight := range statusWeights {
			total += weight
		}
		randNum := rand.Intn(total)
		current := 0
		for status, weight := range statusWeights {
			current += weight
			if randNum < current {
				return status
			}
		}
		return "SUBMITTED" // fallback
	}

	// Helper function to generate score data
	generateScore := func(status string) datatypes.JSON {
		// Only generate scores for SHORTLISTED, OFFERED, and ASSIGNED applications
		// These are applications that have been through screening
		if status != "SHORTLISTED" && status != "OFFERED" && status != "ASSIGNED" {
			return nil
		}

		// Generate realistic scores based on status
		// SHORTLISTED and OFFERED should have good scores (they passed screening)
		// ASSIGNED should have excellent scores (they were selected)
		var autoScore, manualScore, skillMatch, portfolioScore, ratingScore int
		var onTimeRate, reworkRate int

		if status == "ASSIGNED" {
			// Best scores for assigned applications
			autoScore = rand.Intn(15) + 80      // 80-95
			manualScore = rand.Intn(15) + 80    // 80-95
			skillMatch = rand.Intn(15) + 85     // 85-100
			portfolioScore = rand.Intn(20) + 75 // 75-95
			ratingScore = rand.Intn(15) + 85    // 85-100
			onTimeRate = rand.Intn(15) + 85     // 85-100
			reworkRate = rand.Intn(10) + 0      // 0-10 (lower is better)
		} else {
			// Good scores for shortlisted/offered applications
			autoScore = rand.Intn(25) + 65      // 65-90
			manualScore = rand.Intn(20) + 70    // 70-90
			skillMatch = rand.Intn(20) + 75     // 75-95
			portfolioScore = rand.Intn(30) + 60 // 60-90
			ratingScore = rand.Intn(20) + 75    // 75-95
			onTimeRate = rand.Intn(20) + 80     // 80-100
			reworkRate = rand.Intn(15) + 5      // 5-20 (lower is better)
		}

		finalScore := (autoScore + manualScore) / 2

		scoreData := map[string]interface{}{
			"applicationId":         0, // Will be set after creation
			"autoScore":             autoScore,
			"manualSupervisorScore": manualScore,
			"finalScore":            finalScore,
			"skillMatch":            skillMatch,
			"portfolioScore":        portfolioScore,
			"ratingScore":           ratingScore,
			"onTimeRate":            onTimeRate,
			"reworkRate":            reworkRate,
		}

		scoreJSON, _ := json.Marshal(scoreData)
		return datatypes.JSON(scoreJSON)
	}

	for i := 0; i < count; i++ {
		if len(studentUsers) == 0 || len(projects) == 0 {
			continue
		}

		student := studentUsers[rand.Intn(len(studentUsers))]
		proj := projects[rand.Intn(len(projects))]

		// Create student IDs array
		studentIDs := []uint{student.ID}
		studentIDsJSON, _ := json.Marshal(studentIDs)

		status := getWeightedStatus()
		score := generateScore(status)

		app := application.Application{
			ProjectID:     proj.ID,
			ApplicantType: "INDIVIDUAL",
			StudentIDs:    datatypes.JSON(studentIDsJSON),
			Statement:     fmt.Sprintf("I am interested in working on this project because it aligns with my skills and career goals. I have experience in %s and believe I can contribute effectively to the team.", getRandomSkills(3)),
			Status:        status,
			Score:         score,
		}

		// Add OfferExpiresAt for OFFERED applications (7-14 days from now)
		if status == "OFFERED" {
			expiryDays := rand.Intn(7) + 7 // 7-14 days
			expiryDate := time.Now().AddDate(0, 0, expiryDays)
			expiryDateVal := datatypes.Date(expiryDate)
			app.OfferExpiresAt = &expiryDateVal
		}

		db.Create(&app)

		// Update score with application ID if score exists
		if score != nil {
			var scoreData map[string]interface{}
			if err := json.Unmarshal(score, &scoreData); err == nil {
				scoreData["applicationId"] = app.ID
				updatedScoreJSON, _ := json.Marshal(scoreData)
				db.Model(&app).Update("score", datatypes.JSON(updatedScoreJSON))
			}
		}

		applications = append(applications, app)
	}

	return applications
}

// Seed Milestones
func seedMilestones(db *gorm.DB, projects []project.Project) []milestone.Milestone {
	var milestones []milestone.Milestone

	statuses := []string{"PROPOSED", "IN_PROGRESS", "SUBMITTED", "APPROVED", "RELEASED", "COMPLETED"}

	for _, proj := range projects {
		// Create 2-5 milestones per project
		numMilestones := rand.Intn(4) + 2
		for i := 0; i < numMilestones; i++ {
			dueDate := time.Now().AddDate(0, 0, rand.Intn(90)+30).Format("2006-01-02")
			amount := rand.Intn(10000) + 500

			mil := milestone.Milestone{
				ProjectID:          proj.ID,
				Title:              fmt.Sprintf("Milestone %d: %s", i+1, getRandomMilestoneTitle()),
				Scope:              fmt.Sprintf("Complete %s functionality for the project", getRandomFeature()),
				AcceptanceCriteria: fmt.Sprintf("All tests passing, code reviewed, documentation updated"),
				DueDate:            dueDate,
				Amount:             amount,
				Currency:           "USD",
				Status:             statuses[rand.Intn(len(statuses))],
			}
			db.Create(&mil)
			milestones = append(milestones, mil)
		}
	}

	return milestones
}

// Seed Portfolio
func seedPortfolio(db *gorm.DB, users []user.User, projects []project.Project, milestones []milestone.Milestone) []portfolio.PortfolioItem {
	var portfolioItems []portfolio.PortfolioItem

	// Get completed milestones
	completedMilestones := []milestone.Milestone{}
	for _, mil := range milestones {
		if mil.Status == "COMPLETED" || mil.Status == "RELEASED" {
			completedMilestones = append(completedMilestones, mil)
		}
	}

	complexities := []string{"LOW", "MEDIUM", "HIGH"}

	for _, mil := range completedMilestones {
		// Find project
		var proj project.Project
		db.First(&proj, mil.ProjectID)

		// Find students who worked on this project (from applications)
		// Only ASSIGNED applications represent active work
		var apps []application.Application
		db.Where("project_id = ? AND status = ?", mil.ProjectID, "ASSIGNED").Find(&apps)

		for _, app := range apps {
			// Parse student IDs
			var studentIDs []uint
			json.Unmarshal(app.StudentIDs, &studentIDs)

			for _, studentID := range studentIDs {
				var usr user.User
				if db.First(&usr, studentID).Error == nil {
					// Determine complexity
					complexity := complexities[rand.Intn(len(complexities))]
					if mil.Amount > 5000 {
						complexity = "HIGH"
					} else if mil.Amount < 1000 {
						complexity = "LOW"
					}

					// Check if on time
					dueDate, _ := time.Parse("2006-01-02", mil.DueDate)
					onTime := time.Now().Before(dueDate.AddDate(0, 0, 7)) // Within 7 days of due date

					portfolioItem := portfolio.PortfolioItem{
						UserID:          usr.ID,
						ProjectID:       proj.ID,
						MilestoneID:     &mil.ID,
						Role:            "Developer",
						Scope:           mil.Scope,
						Proof:           datatypes.JSON([]byte("[]")),
						Rating:          getRandomRating(),
						Complexity:      complexity,
						AmountDelivered: float64(mil.Amount),
						Currency:        mil.Currency,
						OnTime:          onTime,
						VerifiedAt:      datatypes.Date(time.Now()),
					}
					db.Create(&portfolioItem)
					portfolioItems = append(portfolioItems, portfolioItem)
				}
			}
		}
	}

	return portfolioItems
}

// Seed Groups (student groups for group applications)
func seedGroups(db *gorm.DB, users []user.User, count int) []user.Group {
	var groups []user.Group

	// Get student users
	studentUsers := []user.User{}
	for _, usr := range users {
		if usr.Role == "student" {
			studentUsers = append(studentUsers, usr)
		}
	}

	if len(studentUsers) < 2 {
		return groups
	}

	groupNames := []string{
		"Team Alpha", "Team Beta", "Team Gamma", "Team Delta", "Team Epsilon",
		"Code Warriors", "Tech Titans", "Dev Squad", "Innovation Hub", "Project Masters",
		"Digital Pioneers", "Code Crafters", "Tech Innovators", "Solution Builders", "App Developers",
	}

	for i := 0; i < count && i < len(groupNames); i++ {
		if len(studentUsers) < 2 {
			break
		}

		// Select a leader (first member)
		leader := studentUsers[rand.Intn(len(studentUsers))]

		// Select 1-3 additional members (excluding leader)
		remainingStudents := []user.User{}
		for _, s := range studentUsers {
			if s.ID != leader.ID {
				remainingStudents = append(remainingStudents, s)
			}
		}

		memberCount := rand.Intn(3) + 1 // 1-3 additional members
		if memberCount > len(remainingStudents) {
			memberCount = len(remainingStudents)
		}

		members := []user.User{leader}
		selectedIndices := make(map[int]bool)
		for j := 0; j < memberCount; j++ {
			idx := rand.Intn(len(remainingStudents))
			for selectedIndices[idx] {
				idx = rand.Intn(len(remainingStudents))
			}
			selectedIndices[idx] = true
			members = append(members, remainingStudents[idx])
		}

		group := user.Group{
			UserID:   leader.ID,
			Name:     groupNames[i%len(groupNames)],
			Capacity: len(members) + rand.Intn(3), // Capacity slightly larger than current members
		}

		db.Create(&group)

		// Associate members with group (many-to-many)
		db.Model(&group).Association("Members").Append(members)

		groups = append(groups, group)
	}

	return groups
}

// Seed Chat Messages (messages in groups)
func seedChatMessages(db *gorm.DB, groups []user.Group, users []user.User, count int) []chat.Message {
	var messages []chat.Message

	if len(groups) == 0 {
		return messages
	}

	messageTemplates := []string{
		"Hey everyone, let's discuss the project timeline.",
		"I've completed the initial design mockups.",
		"Can we schedule a meeting this week?",
		"The backend API is ready for testing.",
		"Great work on the frontend components!",
		"Let's review the requirements document together.",
		"I found a bug in the authentication flow.",
		"The database migration is complete.",
		"Should we use React or Vue for this feature?",
		"Let's create a shared document for notes.",
		"Can someone review my pull request?",
		"The deployment is scheduled for tomorrow.",
		"Let's break this into smaller tasks.",
		"I've updated the project documentation.",
		"Thanks for the feedback, I'll make those changes.",
	}

	for i := 0; i < count; i++ {
		group := groups[rand.Intn(len(groups))]

		// Get group members
		var groupWithMembers user.Group
		db.Preload("Members").First(&groupWithMembers, group.ID)

		if len(groupWithMembers.Members) == 0 {
			continue
		}

		sender := groupWithMembers.Members[rand.Intn(len(groupWithMembers.Members))]

		message := chat.Message{
			SenderID: sender.ID,
			GroupID:  group.ID,
			Body:     messageTemplates[rand.Intn(len(messageTemplates))],
		}

		// Randomize message time (within last 30 days)
		daysAgo := rand.Intn(30)
		hoursAgo := rand.Intn(24)
		minutesAgo := rand.Intn(60)
		message.CreatedAt = time.Now().AddDate(0, 0, -daysAgo).Add(-time.Duration(hoursAgo) * time.Hour).Add(-time.Duration(minutesAgo) * time.Minute)

		db.Create(&message)
		messages = append(messages, message)
	}

	return messages
}

// Seed Disputes (disputes between users)
func seedDisputes(db *gorm.DB, users []user.User, count int) []dispute.Dispute {
	var disputes []dispute.Dispute

	subjectTypes := []string{"project", "payment", "milestone", "application", "rating"}
	statuses := []string{"pending", "under_review", "resolved", "dismissed"}
	levels := []string{"low", "medium", "high", "critical"}
	reasons := []string{
		"Payment not received",
		"Work quality dispute",
		"Timeline disagreement",
		"Scope of work mismatch",
		"Communication issues",
		"Deliverable not as specified",
		"Rating dispute",
		"Contract violation",
	}

	resolutions := []string{
		"Resolved through mediation",
		"Partial refund issued",
		"Work to be redone",
		"Dispute dismissed - no violation found",
		"Compensation agreed upon",
		"Terms clarified and accepted",
	}

	for i := 0; i < count; i++ {
		if len(users) < 2 {
			break
		}

		issuer := users[rand.Intn(len(users))]

		// Select a different user as defendant
		defendants := []user.User{}
		for _, u := range users {
			if u.ID != issuer.ID {
				defendants = append(defendants, u)
			}
		}

		if len(defendants) == 0 {
			continue
		}

		defendant := defendants[rand.Intn(len(defendants))]
		status := statuses[rand.Intn(len(statuses))]

		disputeRecord := dispute.Dispute{
			SubjectType: subjectTypes[rand.Intn(len(subjectTypes))],
			Reason:      reasons[rand.Intn(len(reasons))],
			Description: fmt.Sprintf("Dispute regarding %s. %s", subjectTypes[rand.Intn(len(subjectTypes))], reasons[rand.Intn(len(reasons))]),
			Status:      status,
			Level:       levels[rand.Intn(len(levels))],
			IssuerID:    issuer.ID,
			DefendantID: defendant.ID,
		}

		// Add evidence (array of file paths)
		evidenceCount := rand.Intn(3) + 1
		evidence := []string{}
		for j := 0; j < evidenceCount; j++ {
			evidence = append(evidence, fmt.Sprintf("/uploads/disputes/evidence_%d_%d.pdf", disputeRecord.ID, j+1))
		}
		evidenceJSON, _ := json.Marshal(evidence)
		disputeRecord.Evidence = datatypes.JSON(evidenceJSON)

		// If resolved, add resolution details
		if status == "resolved" {
			disputeRecord.Resolution = resolutions[rand.Intn(len(resolutions))]
			daysAgo := rand.Intn(7)
			disputeRecord.ResolvedAt = time.Now().AddDate(0, 0, -daysAgo).Format("2006-01-02")
		}

		db.Create(&disputeRecord)
		disputes = append(disputes, disputeRecord)
	}

	return disputes
}

// Seed Invitations (student/supervisor invitations)
func seedInvitations(db *gorm.DB, organizations []organization.Organization, departments []department.Department, count int) []invitation.Invitation {
	var invitations []invitation.Invitation

	roles := []string{"student", "supervisor"}
	statuses := []string{"PENDING", "USED", "EXPIRED"}

	for i := 0; i < count; i++ {
		org := organizations[rand.Intn(len(organizations))]
		role := roles[rand.Intn(len(roles))]
		status := statuses[rand.Intn(len(statuses))]

		var deptID *uint
		if role == "supervisor" {
			// Find a department in this organization
			orgDepts := []department.Department{}
			for _, dept := range departments {
				if dept.OrganizationID == org.ID {
					orgDepts = append(orgDepts, dept)
				}
			}
			if len(orgDepts) > 0 {
				dept := orgDepts[rand.Intn(len(orgDepts))]
				deptID = &dept.ID
			}
		}

		// Generate token
		token, err := invitation.GenerateToken()
		if err != nil {
			token = fmt.Sprintf("token_%d_%d", time.Now().UnixNano(), i)
		}

		// Generate email
		email := fmt.Sprintf("invite_%s_%d@example.com", role, i)

		inv := invitation.Invitation{
			Email:          email,
			Name:           fmt.Sprintf("%s %s", getRandomFirstName(), getRandomLastName()),
			Role:           role,
			OrganizationID: org.ID,
			DepartmentID:   deptID,
			Token:          token,
			Status:         status,
			ExpiresAt:      time.Now().AddDate(0, 0, 7+rand.Intn(14)), // 7-21 days from now
		}

		// If used, set UsedAt
		if status == "USED" {
			daysAgo := rand.Intn(5)
			usedAt := time.Now().AddDate(0, 0, -daysAgo)
			inv.UsedAt = &usedAt
		}

		// If expired, set ExpiresAt in the past
		if status == "EXPIRED" {
			inv.ExpiresAt = time.Now().AddDate(0, 0, -(rand.Intn(7) + 1))
		}

		db.Create(&inv)
		invitations = append(invitations, inv)
	}

	return invitations
}

// Seed Notifications (user notifications)
func seedNotifications(db *gorm.DB, users []user.User, count int) []notification.Notification {
	var notifications []notification.Notification

	types := []string{
		"application_status", "offer_received", "offer_expiring", "milestone_approved",
		"project_assigned", "message_received", "dispute_created", "invitation_sent",
		"supervisor_request", "payment_received", "deadline_reminder",
	}

	titles := map[string]string{
		"application_status": "Application Status Updated",
		"offer_received":     "New Offer Received",
		"offer_expiring":     "Offer Expiring Soon",
		"milestone_approved": "Milestone Approved",
		"project_assigned":   "Project Assigned",
		"message_received":   "New Message",
		"dispute_created":    "Dispute Created",
		"invitation_sent":    "Invitation Sent",
		"supervisor_request": "Supervisor Request",
		"payment_received":   "Payment Received",
		"deadline_reminder":  "Deadline Reminder",
	}

	messages := map[string]string{
		"application_status": "Your application status has been updated.",
		"offer_received":     "You have received a new project offer.",
		"offer_expiring":     "Your offer will expire in 2 days.",
		"milestone_approved": "Your milestone has been approved.",
		"project_assigned":   "You have been assigned to a new project.",
		"message_received":   "You have a new message in your group.",
		"dispute_created":    "A dispute has been created.",
		"invitation_sent":    "You have been invited to join.",
		"supervisor_request": "A supervisor request has been submitted.",
		"payment_received":   "Payment has been received.",
		"deadline_reminder":  "Project deadline is approaching.",
	}

	links := map[string]string{
		"application_status": "/applications",
		"offer_received":     "/offers",
		"offer_expiring":     "/offers",
		"milestone_approved": "/projects",
		"project_assigned":   "/projects",
		"message_received":   "/messages",
		"dispute_created":    "/disputes",
		"invitation_sent":    "/invitations",
		"supervisor_request": "/supervisor-requests",
		"payment_received":   "/payments",
		"deadline_reminder":  "/projects",
	}

	for i := 0; i < count; i++ {
		user := users[rand.Intn(len(users))]
		notifType := types[rand.Intn(len(types))]

		// Randomize seen status (70% seen, 30% unseen)
		seen := rand.Float32() < 0.7

		notif := notification.Notification{
			Type:    notifType,
			Title:   titles[notifType],
			Message: messages[notifType],
			Seen:    seen,
			Link:    links[notifType],
			UserID:  user.ID,
		}

		// Randomize creation time (within last 30 days)
		daysAgo := rand.Intn(30)
		hoursAgo := rand.Intn(24)
		notif.CreatedAt = time.Now().AddDate(0, 0, -daysAgo).Add(-time.Duration(hoursAgo) * time.Hour)

		db.Create(&notif)
		notifications = append(notifications, notif)
	}

	return notifications
}

// Seed Supervisor Requests (requests from students to supervisors)
func seedSupervisorRequests(db *gorm.DB, projects []project.Project, users []user.User, groups []user.Group, count int) []supervisorrequest.SupervisorRequest {
	var requests []supervisorrequest.SupervisorRequest

	// Get student users
	studentUsers := []user.User{}
	for _, usr := range users {
		if usr.Role == "student" {
			studentUsers = append(studentUsers, usr)
		}
	}

	// Get supervisor users
	supervisorUsers := []user.User{}
	for _, usr := range users {
		if usr.Role == "supervisor" {
			supervisorUsers = append(supervisorUsers, usr)
		}
	}

	if len(studentUsers) == 0 || len(supervisorUsers) == 0 || len(projects) == 0 {
		return requests
	}

	statuses := []string{"PENDING", "APPROVED", "DENIED"}
	messages := []string{
		"I would like to request your supervision for this project.",
		"Could you please supervise our team for this project?",
		"We believe your expertise would be valuable for this project.",
		"Would you be available to supervise this project?",
		"Please consider supervising our work on this project.",
	}

	for i := 0; i < count; i++ {
		project := projects[rand.Intn(len(projects))]
		supervisor := supervisorUsers[rand.Intn(len(supervisorUsers))]

		// Randomly choose between individual student or group
		useGroup := len(groups) > 0 && rand.Float32() < 0.3 // 30% chance of group request

		var studentOrGroupID uint
		if useGroup {
			group := groups[rand.Intn(len(groups))]
			studentOrGroupID = group.ID
		} else {
			student := studentUsers[rand.Intn(len(studentUsers))]
			studentOrGroupID = student.ID
		}

		status := statuses[rand.Intn(len(statuses))]

		request := supervisorrequest.SupervisorRequest{
			ProjectID:        project.ID,
			StudentOrGroupID: studentOrGroupID,
			SupervisorID:     supervisor.ID,
			Status:           status,
			Message:          messages[rand.Intn(len(messages))],
		}

		// Randomize creation time (within last 60 days)
		daysAgo := rand.Intn(60)
		request.CreatedAt = time.Now().AddDate(0, 0, -daysAgo)

		db.Create(&request)
		requests = append(requests, request)
	}

	return requests
}

// Helper functions
func getRandomFirstName() string {
	return ugandanFirstNames[rand.Intn(len(ugandanFirstNames))]
}

func getRandomLastName() string {
	return ugandanLastNames[rand.Intn(len(ugandanLastNames))]
}

func sanitizeEmail(s string) string {
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result += string(c)
		}
	}
	return result
}

func getRandomStreetName() string {
	streets := []string{
		"Kampala Road", "Nakasero Road", "Kololo Hill", "Muyenga Road", "Ntinda Road", "Bukoto Road",
		"Lubowa Road", "Entebbe Road", "Jinja Road", "Bombo Road", "Gayaza Road", "Mukono Road",
		"Masaka Road", "Mbarara Road", "Fort Portal Road", "Arua Road", "Gulu Road", "Lira Road",
		"Mbale Road", "Soroti Road", "Hoima Road", "Masindi Road", "Kabale Road", "Mbarara Road",
		"Kasese Road", "Tororo Road", "Iganga Road", "Jinja Road", "Busia Road", "Malaba Road",
		"Namirembe Road", "Rubaga Road", "Makerere Hill", "Mulago Hill", "Nakawa", "Kireka",
		"Bweyogerere", "Wakiso", "Kajjansi", "Entebbe", "Lubowa", "Seguku", "Buziga", "Munyonyo",
	}
	return streets[rand.Intn(len(streets))]
}

func getRandomSkills(count int) string {
	result := []string{}
	for i := 0; i < count && i < len(skills); i++ {
		result = append(result, skills[rand.Intn(len(skills))])
	}
	return fmt.Sprintf("%v", result)
}

func getRandomMilestoneTitle() string {
	titles := []string{
		"User Authentication",
		"Database Setup",
		"API Development",
		"Frontend Integration",
		"Testing and QA",
		"Deployment",
		"Documentation",
		"Performance Optimization",
		"Security Implementation",
		"Feature Completion",
	}
	return titles[rand.Intn(len(titles))]
}

func getRandomFeature() string {
	features := []string{
		"user management",
		"payment processing",
		"reporting dashboard",
		"notification system",
		"search functionality",
		"data analytics",
		"file upload",
		"messaging system",
		"admin panel",
		"mobile app",
	}
	return features[rand.Intn(len(features))]
}

func getRandomRating() *float64 {
	rating := float64(rand.Intn(3) + 3) // 3.0 to 5.0
	return &rating
}
