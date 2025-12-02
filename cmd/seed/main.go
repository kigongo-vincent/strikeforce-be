package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/BVR-INNOVATION-GROUP/strike-force-backend/config"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// Ugandan data pools
var (
	ugandanFirstNames = []string{
		"James", "John", "Robert", "Michael", "William", "David", "Richard", "Joseph", "Thomas", "Charles",
		"Christopher", "Daniel", "Matthew", "Anthony", "Mark", "Donald", "Steven", "Paul", "Andrew", "Joshua",
		"Kenneth", "Kevin", "Brian", "George", "Timothy", "Ronald", "Edward", "Jason", "Jeffrey", "Ryan",
		"Jacob", "Gary", "Nicholas", "Eric", "Jonathan", "Stephen", "Larry", "Justin", "Scott", "Brandon",
		"Benjamin", "Samuel", "Frank", "Gregory", "Raymond", "Alexander", "Patrick", "Jack", "Dennis", "Jerry",
		"Tyler", "Aaron", "Jose", "Henry", "Adam", "Douglas", "Nathan", "Zachary", "Peter", "Kyle",
		"Noah", "Ethan", "Jeremy", "Walter", "Christian", "Keith", "Roger", "Terry", "Austin", "Sean",
		"Gerald", "Carl", "Harold", "Dylan", "Arthur", "Lawrence", "Jordan", "Jesse", "Bryan", "Billy",
		"Bruce", "Gabriel", "Joe", "Logan", "Alan", "Juan", "Wayne", "Roy", "Ralph", "Randy",
		"Eugene", "Vincent", "Russell", "Louis", "Philip", "Bobby", "Johnny", "Bradley", "Mary", "Patricia",
		"Jennifer", "Linda", "Elizabeth", "Barbara", "Susan", "Jessica", "Sarah", "Karen", "Nancy", "Lisa",
		"Betty", "Margaret", "Sandra", "Ashley", "Kimberly", "Emily", "Donna", "Michelle", "Dorothy", "Carol",
		"Amanda", "Melissa", "Deborah", "Stephanie", "Rebecca", "Sharon", "Laura", "Cynthia", "Kathleen", "Amy",
		"Angela", "Shirley", "Anna", "Brenda", "Pamela", "Emma", "Nicole", "Helen", "Samantha", "Katherine",
		"Christine", "Debra", "Rachel", "Carolyn", "Janet", "Virginia", "Maria", "Heather", "Diane", "Julie",
		"Joyce", "Victoria", "Kelly", "Christina", "Joan", "Evelyn", "Lauren", "Judith", "Megan", "Cheryl",
		"Andrea", "Hannah", "Jacqueline", "Martha", "Gloria", "Teresa", "Sara", "Janice", "Marie", "Julia",
		"Grace", "Judy", "Theresa", "Madison", "Beverly", "Denise", "Marilyn", "Amber", "Danielle", "Rose",
		"Brittany", "Diana", "Abigail", "Jane", "Lori", "Mildred", "Olivia", "Sophia", "Isabella", "Ava",
	}

	ugandanLastNames = []string{
		// Common surnames in Uganda
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
		"Hernandez", "Lopez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee",
		"Thompson", "White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker", "Young",
		"Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores", "Green", "Adams",
		"Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell", "Carter", "Roberts", "Gomez", "Phillips",
		"Evans", "Turner", "Diaz", "Parker", "Cruz", "Edwards", "Collins", "Reyes", "Stewart", "Morris",
		"Morales", "Murphy", "Cook", "Rogers", "Gutierrez", "Ortiz", "Morgan", "Cooper", "Peterson", "Bailey",
		"Reed", "Kelly", "Howard", "Ramos", "Kim", "Cox", "Ward", "Richardson", "Watson", "Brooks",
		"Chavez", "Wood", "James", "Bennett", "Gray", "Mendoza", "Ruiz", "Hughes", "Price", "Alvarez",
		"Castillo", "Sanders", "Patel", "Myers", "Long", "Ross", "Foster", "Jimenez", "Powell", "Jenkins",
		"Perry", "Russell", "Sullivan", "Bell", "Coleman", "Butler", "Henderson", "Barnes", "Gonzales", "Fisher",
		"Vasquez", "Simmons", "Romero", "Jordan", "Patterson", "Alexander", "Hamilton", "Graham", "Reynolds", "Griffin",
		"Wallace", "Moreno", "West", "Cole", "Hayes", "Bryant", "Herrera", "Gibson", "Ellis", "Tran",
		"Medina", "Aguilar", "Stevens", "Murray", "Ford", "Castro", "Marshall", "Owens", "Harrison", "Fernandez",
		"Mcdonald", "Woods", "Washington", "Kennedy", "Wells", "Vargas", "Henry", "Chen", "Freeman", "Webb",
		"Tucker", "Guzman", "Burns", "Crawford", "Olson", "Simpson", "Porter", "Hunter", "Gordon", "Mendez",
		"Silva", "Shaw", "Snyder", "Mason", "Dixon", "Munoz", "Hunt", "Hicks", "Holmes", "Palmer",
		"Wagner", "Black", "Robertson", "Boyd", "Rose", "Stone", "Salazar", "Fox", "Warren", "Mills",
		"Meyer", "Rice", "Schmidt", "Garza", "Daniels", "Ferguson", "Nichols", "Stephens", "Soto", "Weaver",
		"Ryan", "Gardner", "Payne", "Grant", "Dunn", "Kelley", "Spencer", "Hawkins", "Arnold", "Pierce",
		"Vazquez", "Hansen", "Peters", "Santos", "Hart", "Bradley", "Knight", "Elliott", "Cunningham", "Duncan",
		"Armstrong", "Hudson", "Carroll", "Lane", "Riley", "Andrews", "Alvarado", "Ray", "Delgado", "Berry",
		"Perkins", "Hoffman", "Johnston", "Matthews", "Pena", "Richards", "Contreras", "Willis", "Carpenter", "Lawrence",
		"Sandoval", "Guerrero", "George", "Chapman", "Rios", "Estrada", "Ortega", "Watkins", "Greene", "Nunez",
		"Wheeler", "Valdez", "Harper", "Burke", "Larson", "Santiago", "Maldonado", "Morrison", "Franklin", "Carlson",
		"Austin", "Dominguez", "Carr", "Lawson", "Jacobs", "Obrien", "Lynch", "Singh", "Vega", "Bishop",
		"Erickson", "Fletcher", "Mckinney", "Page", "Dawson", "Joseph", "Marquez", "Reeves", "Klein", "Espinoza",
		"Baldwin", "Moran", "Love", "Robbins", "Higgins", "Ball", "Cortez", "Le", "Griffith", "Bowen",
		"Sharp", "Cummings", "Ramsey", "Hardy", "Swanson", "Barber", "Acosta", "Luna", "Chandler", "Daniel",
		"Blair", "Cross", "Simon", "Dennis", "Oconnor", "Quinn", "Gross", "Navarro", "Moss", "Fitzgerald",
		"Doyle", "Mclaughlin", "Rojas", "Rodgers", "Stevenson", "Singh", "Yang", "Figueroa", "Harmon", "Newton",
		"Paul", "Manning", "Garner", "Mcgee", "Reese", "Francis", "Burgess", "Adkins", "Goodman", "Curry",
		"Brady", "Christensen", "Potter", "Walton", "Goodwin", "Mullins", "Molina", "Webster", "Fischer", "Campos",
		"Avila", "Sherman", "Todd", "Chang", "Blake", "Malone", "Wolf", "Hodges", "Juarez", "Gill",
		"Farmer", "Hines", "Gallagher", "Duran", "Hubbard", "Cannon", "Miranda", "Wang", "Saunders", "Tate",
		"Sawyer", "Brooks", "Davenport", "Walters", "Skinner", "Payne", "Gates", "Cohen", "Rogers", "Gregory",
		"Wagner", "Hunter", "Hanson", "Floyd", "Wood", "Gibson", "Mendoza", "Watson", "Hart", "Ray",
		"Oliver", "Lynch", "Graham", "Garza", "Marshall", "Banks", "Santos", "Mccormick", "Patton", "Schneider",
		"Bush", "Thornton", "Mann", "Zimmerman", "Erickson", "Fletcher", "Mckinney", "Page", "Dawson", "Joseph",
		// Additional common surnames
		"Ochieng", "Okello", "Onyango", "Odongo", "Ojok", "Opolot", "Otim", "Okech", "Okot", "Opiyo",
		"Kato", "Mugisha", "Tumusiime", "Nabukeera", "Nakato", "Nalubega", "Nakazibwe", "Nabunya", "Nakayiza", "Nalubwama",
		"Ssemwogerere", "Ssebunya", "Ssemakula", "Ssebowa", "Ssebuliba", "Ssebalamu", "Ssebunya", "Ssemwogerere", "Ssemakula", "Ssebowa",
		"Mukasa", "Mutebi", "Mugerwa", "Mugabi", "Mugisha", "Mutebi", "Mukasa", "Mugerwa", "Mugabi", "Mugisha",
		"Kigozi", "Kisitu", "Kigozi", "Kisitu", "Kigozi", "Kisitu", "Kigozi", "Kisitu", "Kigozi", "Kisitu",
	}

	ugandanUniversities = []string{
		"Makerere University",
		"Kyambogo University",
		"Uganda Christian University",
		"Uganda Martyrs University",
		"Mbarara University of Science and Technology",
		"Gulu University",
		"Busitema University",
		"Kabale University",
		"Lira University",
		"Muni University",
		"Bishop Stuart University",
		"Bugema University",
		"Islamic University in Uganda",
		"Kampala International University",
		"Mountains of the Moon University",
		"Ndejje University",
		"Uganda Management Institute",
		"Victoria University",
		"International University of East Africa",
		"St. Lawrence University",
	}

	ugandanCompanies = []string{
		"MTN Uganda",
		"Airtel Uganda",
		"Centenary Bank",
		"Stanbic Bank Uganda",
		"Equity Bank Uganda",
		"Bank of Uganda",
		"Uganda Revenue Authority",
		"National Water and Sewerage Corporation",
		"Uganda Electricity Distribution Company",
		"Uganda National Roads Authority",
		"Kakira Sugar Works",
		"Bidco Uganda",
		"Mukwano Industries",
		"Rwenzori Bottling Company",
		"Coca-Cola Beverages Africa",
		"Uganda Breweries",
		"Century Bottling Company",
		"Jumia Uganda",
		"SafeBoda",
		"Tugende",
		"Ensibuuko",
		"Laboremus Uganda",
		"Fundi Bots",
		"Reflex Energy",
		"SolarNow",
		"BRAC Uganda",
		"World Vision Uganda",
		"Save the Children Uganda",
		"Oxfam Uganda",
		"Plan International Uganda",
		"ActionAid Uganda",
		"Amref Health Africa",
		"Uganda Red Cross Society",
		"Uganda Cancer Institute",
		"Mulago National Referral Hospital",
		"Kampala Capital City Authority",
		"National Agricultural Research Organisation",
		"Uganda National Bureau of Standards",
		"Uganda Investment Authority",
	}

	ugandanDepartments = []string{
		"Computer Science",
		"Information Technology",
		"Software Engineering",
		"Electrical Engineering",
		"Mechanical Engineering",
		"Civil Engineering",
		"Business Administration",
		"Accounting",
		"Finance",
		"Marketing",
		"Economics",
		"Statistics",
		"Mathematics",
		"Physics",
		"Chemistry",
		"Biology",
		"Agriculture",
		"Environmental Science",
		"Public Health",
		"Medicine",
		"Nursing",
		"Education",
		"Law",
		"Journalism",
		"Social Work",
		"Psychology",
		"Sociology",
		"Political Science",
		"International Relations",
		"Linguistics",
	}

	ugandanCourses = []string{
		"Bachelor of Science in Computer Science",
		"Bachelor of Science in Information Technology",
		"Bachelor of Science in Software Engineering",
		"Bachelor of Science in Electrical Engineering",
		"Bachelor of Science in Mechanical Engineering",
		"Bachelor of Science in Civil Engineering",
		"Bachelor of Business Administration",
		"Bachelor of Commerce",
		"Bachelor of Accounting",
		"Bachelor of Finance",
		"Bachelor of Economics",
		"Bachelor of Statistics",
		"Bachelor of Science in Mathematics",
		"Bachelor of Science in Physics",
		"Bachelor of Science in Chemistry",
		"Bachelor of Science in Biology",
		"Bachelor of Agriculture",
		"Bachelor of Environmental Science",
		"Bachelor of Public Health",
		"Bachelor of Medicine and Bachelor of Surgery",
		"Bachelor of Science in Nursing",
		"Bachelor of Education",
		"Bachelor of Laws",
		"Bachelor of Journalism",
		"Bachelor of Social Work",
		"Bachelor of Arts in Psychology",
		"Bachelor of Arts in Sociology",
		"Bachelor of Arts in Political Science",
		"Bachelor of Arts in International Relations",
		"Bachelor of Arts in Linguistics",
	}

	projectTitles = []string{
		"E-Commerce Platform Development",
		"Mobile Banking Application",
		"School Management System",
		"Hospital Management System",
		"Agricultural Market Platform",
		"Transportation Booking System",
		"Real Estate Management System",
		"Library Management System",
		"Hotel Booking System",
		"Restaurant Management System",
		"Online Learning Platform",
		"Job Portal Development",
		"Social Media Application",
		"Healthcare Information System",
		"Supply Chain Management System",
		"Financial Management System",
		"Human Resource Management System",
		"Inventory Management System",
		"Customer Relationship Management System",
		"Document Management System",
		"Event Management Platform",
		"Farming Management System",
		"Water Management System",
		"Energy Monitoring System",
		"Waste Management System",
		"Tourism Booking Platform",
		"Insurance Management System",
		"Legal Case Management System",
		"Research Data Management System",
		"Community Engagement Platform",
	}

	projectDescriptions = []string{
		"A comprehensive platform for managing online transactions and inventory",
		"Secure mobile application for banking services",
		"Complete system for managing school operations and student records",
		"Integrated system for hospital operations and patient management",
		"Digital platform connecting farmers with buyers",
		"System for booking and managing transportation services",
		"Platform for managing real estate properties and transactions",
		"Digital system for library operations and book management",
		"Online platform for hotel reservations and management",
		"System for managing restaurant operations and orders",
		"E-learning platform with course management features",
		"Job portal connecting employers with job seekers",
		"Social networking application with messaging features",
		"Healthcare system for patient records and appointments",
		"System for managing supply chain operations",
		"Financial system for accounting and reporting",
		"HR system for employee management and payroll",
		"Inventory tracking and management system",
		"CRM system for customer relationship management",
		"Document storage and management system",
		"Platform for organizing and managing events",
		"System for managing farming operations",
		"Water resource management and monitoring system",
		"Energy consumption monitoring and management",
		"Waste collection and management system",
		"Tourism booking and management platform",
		"Insurance policy and claim management system",
		"Legal case tracking and management system",
		"Research data collection and analysis system",
		"Community engagement and communication platform",
	}

	skills = []string{
		"JavaScript", "Python", "Java", "C++", "C#", "PHP", "Ruby", "Go", "Swift", "Kotlin",
		"React", "Angular", "Vue.js", "Node.js", "Express", "Django", "Flask", "Spring", "Laravel", "Rails",
		"MySQL", "PostgreSQL", "MongoDB", "Redis", "SQLite", "Oracle", "SQL Server",
		"HTML", "CSS", "SASS", "Bootstrap", "Tailwind CSS",
		"Git", "Docker", "Kubernetes", "AWS", "Azure", "GCP",
		"REST API", "GraphQL", "Microservices", "CI/CD", "DevOps",
		"Machine Learning", "Data Science", "Artificial Intelligence",
		"Mobile Development", "iOS", "Android", "React Native", "Flutter",
		"Project Management", "Agile", "Scrum", "Kanban",
		"UI/UX Design", "Figma", "Adobe XD", "Sketch",
		"Testing", "Jest", "JUnit", "Selenium", "Cypress",
		"Linux", "Windows Server", "Network Administration",
		"Cybersecurity", "Penetration Testing", "Ethical Hacking",
		"Blockchain", "Smart Contracts", "Cryptocurrency",
		"Data Analytics", "Business Intelligence", "Tableau", "Power BI",
	}
)

func main() {
	// Load environment variables from .env file
	envErr := godotenv.Load()
	if envErr != nil {
		// Try loading from parent directory (backend/.env)
		envErr = godotenv.Load("../.env")
		if envErr != nil {
			log.Println("Warning: Failed to load .env file. Make sure environment variables are set.")
		}
	}

	// Parse command line arguments
	var numOrganizations, numUsers, numProjects, numApplications, numGroups, numMessages, numDisputes, numInvitations, numNotifications, numSupervisorRequests int
	if len(os.Args) > 1 {
		if n, err := strconv.Atoi(os.Args[1]); err == nil {
			numOrganizations = n
		}
	}
	if len(os.Args) > 2 {
		if n, err := strconv.Atoi(os.Args[2]); err == nil {
			numUsers = n
		}
	}
	if len(os.Args) > 3 {
		if n, err := strconv.Atoi(os.Args[3]); err == nil {
			numProjects = n
		}
	}
	if len(os.Args) > 4 {
		if n, err := strconv.Atoi(os.Args[4]); err == nil {
			numApplications = n
		}
	}
	if len(os.Args) > 5 {
		if n, err := strconv.Atoi(os.Args[5]); err == nil {
			numGroups = n
		}
	}
	if len(os.Args) > 6 {
		if n, err := strconv.Atoi(os.Args[6]); err == nil {
			numMessages = n
		}
	}
	if len(os.Args) > 7 {
		if n, err := strconv.Atoi(os.Args[7]); err == nil {
			numDisputes = n
		}
	}
	if len(os.Args) > 8 {
		if n, err := strconv.Atoi(os.Args[8]); err == nil {
			numInvitations = n
		}
	}
	if len(os.Args) > 9 {
		if n, err := strconv.Atoi(os.Args[9]); err == nil {
			numNotifications = n
		}
	}
	if len(os.Args) > 10 {
		if n, err := strconv.Atoi(os.Args[10]); err == nil {
			numSupervisorRequests = n
		}
	}

	// Default values
	if numOrganizations == 0 {
		numOrganizations = 10
	}
	if numUsers == 0 {
		numUsers = 100
	}
	if numProjects == 0 {
		numProjects = 50
	}
	if numApplications == 0 {
		numApplications = 200
	}
	if numGroups == 0 {
		numGroups = 20
	}
	if numMessages == 0 {
		numMessages = 500
	}
	if numDisputes == 0 {
		numDisputes = 30
	}
	if numInvitations == 0 {
		numInvitations = 50
	}
	if numNotifications == 0 {
		numNotifications = 300
	}
	if numSupervisorRequests == 0 {
		numSupervisorRequests = 40
	}

	fmt.Printf("Starting database seeding...\n")
	fmt.Printf("Organizations: %d\n", numOrganizations)
	fmt.Printf("Users: %d\n", numUsers)
	fmt.Printf("Projects: %d\n", numProjects)
	fmt.Printf("Applications: %d\n", numApplications)
	fmt.Printf("Groups: %d\n", numGroups)
	fmt.Printf("Chat Messages: %d\n", numMessages)
	fmt.Printf("Disputes: %d\n", numDisputes)
	fmt.Printf("Invitations: %d\n", numInvitations)
	fmt.Printf("Notifications: %d\n", numNotifications)
	fmt.Printf("Supervisor Requests: %d\n", numSupervisorRequests)

	// Connect to database
	db, err := config.ConnectToDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Seed data
	rand.Seed(time.Now().UnixNano())

	// Hash password for all users
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("1234567890"), bcrypt.DefaultCost)
	passwordHash := string(hashedPassword)

	// Seed Organizations
	fmt.Println("\nSeeding organizations...")
	organizations := seedOrganizations(db, numOrganizations, passwordHash)
	fmt.Printf("Created %d organizations\n", len(organizations))

	// Seed Departments and Courses
	fmt.Println("\nSeeding departments and courses...")
	departments := seedDepartments(db, organizations)
	courses := seedCourses(db, departments)
	fmt.Printf("Created %d departments and %d courses\n", len(departments), len(courses))

	// Seed Users (students, supervisors, partners, university-admins)
	fmt.Println("\nSeeding users...")
	users := seedUsers(db, organizations, courses, passwordHash, numUsers)
	fmt.Printf("Created %d users\n", len(users))

	// Seed Students
	fmt.Println("\nSeeding students...")
	students := seedStudents(db, users, courses)
	fmt.Printf("Created %d student records\n", len(students))

	// Seed Supervisors
	fmt.Println("\nSeeding supervisors...")
	supervisors := seedSupervisors(db, users, departments)
	fmt.Printf("Created %d supervisor records\n", len(supervisors))

	// Seed Projects
	fmt.Println("\nSeeding projects...")
	projects := seedProjects(db, users, departments, supervisors, numProjects)
	fmt.Printf("Created %d projects\n", len(projects))

	// Seed Applications
	fmt.Println("\nSeeding applications...")
	applications := seedApplications(db, users, projects, numApplications)
	fmt.Printf("Created %d applications\n", len(applications))

	// Seed Milestones
	fmt.Println("\nSeeding milestones...")
	milestones := seedMilestones(db, projects)
	fmt.Printf("Created %d milestones\n", len(milestones))

	// Seed Portfolio Items
	fmt.Println("\nSeeding portfolio items...")
	portfolioItems := seedPortfolio(db, users, projects, milestones)
	fmt.Printf("Created %d portfolio items\n", len(portfolioItems))

	// Seed Groups
	fmt.Println("\nSeeding groups...")
	groups := seedGroups(db, users, numGroups)
	fmt.Printf("Created %d groups\n", len(groups))

	// Seed Chat Messages
	fmt.Println("\nSeeding chat messages...")
	messages := seedChatMessages(db, groups, users, numMessages)
	fmt.Printf("Created %d chat messages\n", len(messages))

	// Seed Disputes
	fmt.Println("\nSeeding disputes...")
	disputes := seedDisputes(db, users, numDisputes)
	fmt.Printf("Created %d disputes\n", len(disputes))

	// Seed Invitations
	fmt.Println("\nSeeding invitations...")
	invitations := seedInvitations(db, organizations, departments, numInvitations)
	fmt.Printf("Created %d invitations\n", len(invitations))

	// Seed Notifications
	fmt.Println("\nSeeding notifications...")
	notifications := seedNotifications(db, users, numNotifications)
	fmt.Printf("Created %d notifications\n", len(notifications))

	// Seed Supervisor Requests
	fmt.Println("\nSeeding supervisor requests...")
	supervisorRequests := seedSupervisorRequests(db, projects, users, groups, numSupervisorRequests)
	fmt.Printf("Created %d supervisor requests\n", len(supervisorRequests))

	fmt.Println("\n✅ Database seeding completed successfully!")
	fmt.Printf("\nSummary:")
	fmt.Printf("\n  - Organizations: %d", len(organizations))
	fmt.Printf("\n  - Departments: %d", len(departments))
	fmt.Printf("\n  - Courses: %d", len(courses))
	fmt.Printf("\n  - Users: %d", len(users))
	fmt.Printf("\n  - Students: %d", len(students))
	fmt.Printf("\n  - Supervisors: %d", len(supervisors))
	fmt.Printf("\n  - Projects: %d", len(projects))
	fmt.Printf("\n  - Applications: %d", len(applications))
	fmt.Printf("\n  - Milestones: %d", len(milestones))
	fmt.Printf("\n  - Portfolio Items: %d", len(portfolioItems))
	fmt.Printf("\n  - Groups: %d", len(groups))
	fmt.Printf("\n  - Chat Messages: %d", len(messages))
	fmt.Printf("\n  - Disputes: %d", len(disputes))
	fmt.Printf("\n  - Invitations: %d", len(invitations))
	fmt.Printf("\n  - Notifications: %d", len(notifications))
	fmt.Printf("\n  - Supervisor Requests: %d\n", len(supervisorRequests))
}
