package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	var (
		databaseURL string
		email       string
		password    string
		name        string
	)

	flag.StringVar(&databaseURL, "database", "", "Database URL")
	flag.StringVar(&email, "email", "", "Admin email")
	flag.StringVar(&password, "password", "", "Admin password")
	flag.StringVar(&name, "name", "Admin", "Admin name")
	flag.Parse()

	// Fall back to env vars
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
	}
	if email == "" {
		email = os.Getenv("ADMIN_EMAIL")
	}
	if password == "" {
		password = os.Getenv("ADMIN_PASSWORD")
	}
	if name == "" || name == "Admin" {
		if n := os.Getenv("ADMIN_NAME"); n != "" {
			name = n
		}
	}

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if email == "" {
		log.Fatal("ADMIN_EMAIL is required")
	}
	if password == "" {
		log.Fatal("ADMIN_PASSWORD is required")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Seed roles
	fmt.Println("Seeding roles...")
	if err := seedRoles(ctx, conn); err != nil {
		log.Fatalf("Failed to seed roles: %v", err)
	}
	fmt.Println("Roles seeded successfully")

	// Get admin role ID
	var adminRoleID string
	err = conn.QueryRow(ctx, `SELECT id FROM roles WHERE slug = 'admin'`).Scan(&adminRoleID)
	if err != nil {
		log.Fatalf("Failed to get admin role: %v", err)
	}

	// Upsert admin user with role_id
	_, err = conn.Exec(ctx, `
		INSERT INTO users (email, password_hash, name, role_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			name = EXCLUDED.name,
			role_id = EXCLUDED.role_id
	`, email, string(hash), name, adminRoleID)

	if err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("Super admin user created/updated: %s\n", email)

	// Also create corresponding author for account page (marked as system user - cannot be deleted)
	slug := generateSlug(name)
	_, err = conn.Exec(ctx, `
		INSERT INTO authors (name, slug, email, role_id, is_system)
		VALUES ($1, $2, $3, $4, true)
		ON CONFLICT (email) DO UPDATE SET
			name = EXCLUDED.name,
			role_id = EXCLUDED.role_id,
			is_system = true
	`, name, slug, email, adminRoleID)

	if err != nil {
		log.Fatalf("Failed to create super admin author profile: %v", err)
	}

	fmt.Printf("Super admin author profile created/updated: %s (protected from deletion)\n", email)

	// Seed categories
	fmt.Println("Seeding categories...")
	if err := seedCategories(ctx, conn); err != nil {
		log.Fatalf("Failed to seed categories: %v", err)
	}
	fmt.Println("Categories seeded successfully")

	// Seed tags
	fmt.Println("Seeding tags...")
	if err := seedTags(ctx, conn); err != nil {
		log.Fatalf("Failed to seed tags: %v", err)
	}
	fmt.Println("Tags seeded successfully")

	// Seed sample articles
	fmt.Println("Seeding articles...")
	if err := seedArticles(ctx, conn, email); err != nil {
		log.Fatalf("Failed to seed articles: %v", err)
	}
	fmt.Println("Articles seeded successfully")

	fmt.Println("\n✓ Database seeding completed!")
}

func seedRoles(ctx context.Context, conn *pgx.Conn) error {
	// Seed the three main roles
	roles := []struct {
		name        string
		slug        string
		description string
		isSystem    bool
	}{
		{"Administrator", "admin", "Full access to all features and settings", true},
		{"Author", "author", "Can manage articles, categories, and tags", true},
		{"User", "user", "Basic user with limited access", true},
	}

	for _, role := range roles {
		_, err := conn.Exec(ctx, `
			INSERT INTO roles (name, slug, description, is_system)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (slug) DO NOTHING
		`, role.name, role.slug, role.description, role.isSystem)
		if err != nil {
			return fmt.Errorf("failed to seed role %s: %w", role.slug, err)
		}
		fmt.Printf("  - Role '%s' seeded\n", role.name)
	}

	return nil
}

func seedCategories(ctx context.Context, conn *pgx.Conn) error {
	categories := []struct {
		name        string
		slug        string
		description string
	}{
		{"National Politics", "national-politics", "Coverage of Malacañang, Congress, and national government affairs including executive orders, legislative proceedings, and inter-agency coordination"},
		{"Local Government", "local-government", "News from provinces, cities, municipalities, and barangays including LGU programs, local ordinances, and regional development initiatives"},
		{"Elections", "elections", "COMELEC updates, campaign coverage, candidate profiles, election results, and voter education for national and local elections"},
		{"Policy", "policy", "Analysis of government policies, programs, and reforms including economic measures, social welfare initiatives, and regulatory changes"},
		{"Opinion", "opinion", "Editorial commentary, expert analysis, and diverse perspectives on Philippine political issues and governance"},
		{"Investigations", "investigations", "In-depth investigative journalism on corruption, anomalies, and accountability in government"},
		{"Legislative", "legislative", "Bills, laws, committee hearings, and proceedings from the Senate and House of Representatives"},
		{"Judiciary", "judiciary", "Supreme Court decisions, legal cases involving public officials, and justice system developments"},
	}

	for _, cat := range categories {
		_, err := conn.Exec(ctx, `
			INSERT INTO categories (name, slug, description)
			VALUES ($1, $2, $3)
			ON CONFLICT (slug) DO NOTHING
		`, cat.name, cat.slug, cat.description)
		if err != nil {
			return fmt.Errorf("failed to seed category %s: %w", cat.slug, err)
		}
		fmt.Printf("  - Category '%s' seeded\n", cat.name)
	}

	return nil
}

func seedTags(ctx context.Context, conn *pgx.Conn) error {
	tags := []struct {
		name string
		slug string
	}{
		// === Content Types ===
		{"Breaking News", "breaking-news"},
		{"Analysis", "analysis"},
		{"Interview", "interview"},
		{"Feature", "feature"},
		{"Editorial", "editorial"},
		{"Fact Check", "fact-check"},
		{"Exclusive", "exclusive"},

		// === Government Institutions ===
		{"Malacañang", "malacanang"},
		{"Senate", "senate"},
		{"House of Representatives", "house-of-representatives"},
		{"Supreme Court", "supreme-court"},
		{"Ombudsman", "ombudsman"},
		{"Commission on Audit", "coa"},
		{"Civil Service Commission", "csc"},
		{"COMELEC", "comelec"},
		{"Commission on Human Rights", "chr"},

		// === Executive Departments ===
		{"DOJ", "doj"},
		{"DILG", "dilg"},
		{"DFA", "dfa"},
		{"DOF", "dof"},
		{"DBM", "dbm"},
		{"DOLE", "dole"},
		{"DENR", "denr"},
		{"DOH", "doh"},
		{"DepEd", "deped"},
		{"DSWD", "dswd"},
		{"DOTr", "dotr"},
		{"DPWH", "dpwh"},
		{"DA", "da"},
		{"DTI", "dti"},
		{"DICT", "dict"},

		// === Key Political Topics ===
		{"National Budget", "national-budget"},
		{"Taxation", "taxation"},
		{"Public Debt", "public-debt"},
		{"Inflation", "inflation"},
		{"Charter Change", "charter-change"},
		{"Federalism", "federalism"},
		{"Political Dynasty", "political-dynasty"},
		{"Party-list", "party-list"},
		{"Electoral Reform", "electoral-reform"},

		// === Foreign Relations ===
		{"West Philippine Sea", "west-philippine-sea"},
		{"China Relations", "china-relations"},
		{"US Relations", "us-relations"},
		{"ASEAN", "asean"},
		{"Foreign Policy", "foreign-policy"},
		{"VFA", "vfa"},
		{"EDCA", "edca"},

		// === Security & Defense ===
		{"AFP", "afp"},
		{"PNP", "pnp"},
		{"NBI", "nbi"},
		{"Drug War", "drug-war"},
		{"Insurgency", "insurgency"},
		{"NPA", "npa"},
		{"Terrorism", "terrorism"},
		{"Peace Process", "peace-process"},
		{"Bangsamoro", "bangsamoro"},
		{"BARMM", "barmm"},

		// === Governance Issues ===
		{"Corruption", "corruption"},
		{"Transparency", "transparency"},
		{"FOI", "foi"},
		{"Accountability", "accountability"},
		{"Good Governance", "good-governance"},
		{"Red Tape", "red-tape"},
		{"SALN", "saln"},

		// === Social Issues ===
		{"Poverty", "poverty"},
		{"Education", "education"},
		{"Healthcare", "healthcare"},
		{"Housing", "housing"},
		{"OFW", "ofw"},
		{"Labor", "labor"},
		{"Minimum Wage", "minimum-wage"},
		{"Social Welfare", "social-welfare"},
		{"4Ps", "4ps"},

		// === Infrastructure & Development ===
		{"Infrastructure", "infrastructure"},
		{"Build Better More", "build-better-more"},
		{"PPP", "ppp"},
		{"Transportation", "transportation"},
		{"MRT/LRT", "mrt-lrt"},
		{"NLEX/SLEX", "nlex-slex"},

		// === Economy ===
		{"Economy", "economy"},
		{"GDP", "gdp"},
		{"Investment", "investment"},
		{"PEZA", "peza"},
		{"BSP", "bsp"},
		{"PSE", "pse"},
		{"Agriculture", "agriculture"},
		{"Rice Tariffication", "rice-tariffication"},

		// === Human Rights & Justice ===
		{"Human Rights", "human-rights"},
		{"EJK", "ejk"},
		{"Press Freedom", "press-freedom"},
		{"ICC", "icc"},
		{"Political Prisoners", "political-prisoners"},
		{"Martial Law", "martial-law"},

		// === Environment ===
		{"Climate Change", "climate-change"},
		{"Mining", "mining"},
		{"Disaster Response", "disaster-response"},
		{"NDRRMC", "ndrrmc"},
	}

	for _, tag := range tags {
		_, err := conn.Exec(ctx, `
			INSERT INTO tags (name, slug)
			VALUES ($1, $2)
			ON CONFLICT (slug) DO NOTHING
		`, tag.name, tag.slug)
		if err != nil {
			return fmt.Errorf("failed to seed tag %s: %w", tag.slug, err)
		}
		fmt.Printf("  - Tag '%s' seeded\n", tag.name)
	}

	return nil
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	reg := regexp.MustCompile("[^a-z0-9-]")
	slug = reg.ReplaceAllString(slug, "")
	return slug
}

func seedArticles(ctx context.Context, conn *pgx.Conn, authorEmail string) error {
	// Get author ID by email
	var authorID string
	err := conn.QueryRow(ctx, `SELECT id FROM authors WHERE email = $1`, authorEmail).Scan(&authorID)
	if err != nil {
		return fmt.Errorf("failed to get author ID: %w", err)
	}

	// Get category IDs
	categoryIDs := make(map[string]string)
	rows, err := conn.Query(ctx, `SELECT slug, id FROM categories`)
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var slug, id string
		if err := rows.Scan(&slug, &id); err != nil {
			return err
		}
		categoryIDs[slug] = id
	}

	// Get tag IDs
	tagIDs := make(map[string]string)
	tagRows, err := conn.Query(ctx, `SELECT slug, id FROM tags`)
	if err != nil {
		return fmt.Errorf("failed to get tags: %w", err)
	}
	defer tagRows.Close()
	for tagRows.Next() {
		var slug, id string
		if err := tagRows.Scan(&slug, &id); err != nil {
			return err
		}
		tagIDs[slug] = id
	}

	articles := []struct {
		slug         string
		title        string
		summary      string
		content      string
		categorySlug string
		tagSlugs     []string
	}{
		{
			slug:    "marcos-administration-strengthens-west-philippine-sea-defense-strategy",
			title:   "Marcos Administration Strengthens West Philippine Sea Defense Strategy with Enhanced Maritime Patrols",
			summary: "The Philippine government announces expanded maritime operations and diplomatic initiatives to protect sovereignty in the West Philippine Sea amid ongoing territorial disputes.",
			content: `<p>President Ferdinand Marcos Jr. has announced a comprehensive strategy to strengthen the Philippines' position in the West Philippine Sea, combining enhanced maritime patrols with intensified diplomatic engagement.</p>

<h2>Military Modernization Efforts</h2>
<p>The Armed Forces of the Philippines (AFP) has begun deploying additional patrol vessels to contested waters, backed by the recent acquisition of new maritime surveillance equipment. Defense Secretary Gilberto Teodoro emphasized that these measures are purely defensive in nature.</p>

<blockquote>"We are simply exercising our sovereign rights within our exclusive economic zone. Our fishermen deserve protection, and our territory deserves defense," Teodoro stated during a press briefing at Camp Aguinaldo.</blockquote>

<h2>Diplomatic Front</h2>
<p>The Department of Foreign Affairs (DFA) has filed multiple diplomatic protests following recent incidents involving foreign vessels in Philippine waters. Secretary Enrique Manalo has been conducting a series of bilateral meetings with ASEAN counterparts to build regional consensus on the South China Sea issue.</p>

<h2>International Support</h2>
<p>The administration has garnered support from key allies, with the United States reaffirming its commitment under the Mutual Defense Treaty. Japan and Australia have also expressed solidarity with the Philippines' position, with joint naval exercises scheduled for the coming months.</p>

<h2>Economic Implications</h2>
<p>Experts note that securing Philippine waters is crucial for the country's fishing industry and potential energy exploration. The Department of Energy has indicated interest in resuming oil and gas exploration projects in the Reed Bank area, pending security assessments.</p>`,
			categorySlug: "national-politics",
			tagSlugs:     []string{"west-philippine-sea", "afp", "dfa", "foreign-policy", "us-relations", "breaking-news"},
		},
		{
			slug:    "2025-midterm-elections-comelec-announces-final-candidate-list",
			title:   "2025 Midterm Elections: COMELEC Releases Final List of Senatorial Candidates",
			summary: "The Commission on Elections publishes the official roster of candidates for the 2025 midterm elections, featuring political veterans and newcomers vying for 12 Senate seats.",
			content: `<p>The Commission on Elections (COMELEC) has released the final list of candidates for the 2025 midterm elections, setting the stage for what analysts predict will be a highly competitive race for 12 Senate seats.</p>

<h2>Key Candidates</h2>
<p>The roster includes several administration-backed candidates, opposition figures, and independent aspirants. Notable names include former cabinet secretaries, incumbent local officials, and celebrities entering the political arena for the first time.</p>

<h2>Party-List Registration</h2>
<p>COMELEC also announced that 177 party-list groups have been accredited for the upcoming elections. The commission noted stricter verification processes have been implemented to ensure groups genuinely represent marginalized sectors.</p>

<blockquote>"We have learned from past elections. This time, we're implementing more rigorous background checks and requiring stronger proof of genuine advocacy," COMELEC Chairman George Garcia explained.</blockquote>

<h2>Campaign Period</h2>
<p>The official campaign period for national positions will begin on February 12, 2025, while local campaigns will start on March 28. COMELEC has reminded candidates of spending limits and has deployed field monitors nationwide.</p>

<h2>Election Technology</h2>
<p>The commission confirmed that it will continue using automated counting machines while implementing additional security protocols. Random manual audits will be conducted in more precincts compared to previous elections to enhance transparency.</p>

<h2>Voter Registration</h2>
<p>Final voter registration figures show approximately 67 million registered voters, an increase of 4 million from the 2022 elections. Youth voters (18-30) comprise roughly 30% of the electorate.</p>`,
			categorySlug: "elections",
			tagSlugs:     []string{"comelec", "electoral-reform", "party-list", "analysis"},
		},
		{
			slug:    "senate-approves-maharlika-investment-fund-amendments",
			title:   "Senate Approves Amendments to Maharlika Investment Fund for Enhanced Oversight",
			summary: "The upper chamber passes amendments requiring stricter reporting requirements and independent audits for the sovereign wealth fund.",
			content: `<p>The Philippine Senate has approved amendments to Republic Act No. 11954, or the Maharlika Investment Fund (MIF) Law, introducing enhanced oversight mechanisms following calls for greater transparency in the management of the sovereign wealth fund.</p>

<h2>Key Amendments</h2>
<p>The approved changes include mandatory quarterly reporting to Congress, independent third-party audits, and stricter conflict of interest provisions for fund managers. Senator Grace Poe, chair of the Senate Committee on Finance, sponsored the amendments.</p>

<blockquote>"The Filipino people have entrusted their resources to this fund. They deserve nothing less than complete transparency in how these investments are managed," Senator Poe said during floor deliberations.</blockquote>

<h2>Investment Guidelines</h2>
<p>The amendments also clarify investment priorities, emphasizing infrastructure projects, renewable energy, and technology ventures. At least 40% of investments must be directed toward domestic projects that create local employment.</p>

<h2>Opposition Concerns</h2>
<p>Some opposition senators had pushed for even stricter measures, including requiring Congressional approval for investments exceeding PHP 10 billion. These proposals were not included in the final version but may be revisited in future sessions.</p>

<h2>Implementation Timeline</h2>
<p>The Maharlika Investment Corporation (MIC) board will have 90 days from the law's effectivity to implement the new reporting framework. The first comprehensive audit under the new guidelines is expected by Q3 2025.</p>`,
			categorySlug: "legislative",
			tagSlugs:     []string{"senate", "national-budget", "transparency", "good-governance", "investment"},
		},
		{
			slug:    "dilg-launches-local-government-performance-index",
			title:   "DILG Launches New Performance Index to Rank Local Government Units",
			summary: "The Department of the Interior and Local Government introduces a comprehensive ranking system measuring LGU efficiency, transparency, and citizen satisfaction.",
			content: `<p>The Department of the Interior and Local Government (DILG) has unveiled the Local Government Performance Index (LGPI), a new system designed to evaluate and rank the performance of cities, municipalities, and provinces across the country.</p>

<h2>Evaluation Criteria</h2>
<p>The index assesses LGUs across five key areas: public service delivery, fiscal management, transparency and accountability, citizen engagement, and disaster preparedness. Each category carries equal weight in the overall score.</p>

<h2>Incentives and Consequences</h2>
<p>DILG Secretary Benhur Abalos announced that high-performing LGUs will receive priority consideration for national government projects and additional Internal Revenue Allotment (IRA) incentives. Consistently underperforming units will face administrative reviews.</p>

<blockquote>"This is not about punishing local officials. It's about setting standards and encouraging a culture of excellence in local governance," Secretary Abalos clarified during the launch ceremony.</blockquote>

<h2>Data-Driven Approach</h2>
<p>The LGPI will utilize data from multiple sources, including the Commission on Audit, civil society organizations, and citizen satisfaction surveys. Third-party validators will ensure the accuracy and integrity of assessments.</p>

<h2>Initial Rankings</h2>
<p>Preliminary assessments covering 2024 performance will be released in March 2025. The full index, including all 1,700+ LGUs, will be published annually starting in 2026.</p>

<h2>LGU Reactions</h2>
<p>The League of Cities of the Philippines and League of Municipalities have expressed general support for the initiative while requesting additional capacity-building assistance for LGUs facing resource constraints.</p>`,
			categorySlug: "local-government",
			tagSlugs:     []string{"dilg", "good-governance", "transparency", "feature"},
		},
		{
			slug:    "house-committee-investigates-philhealth-fund-management",
			title:   "House Committee Opens Investigation into PhilHealth Fund Management Practices",
			summary: "Congressional inquiry examines allegations of fund mismanagement and delayed reimbursements affecting healthcare providers nationwide.",
			content: `<p>The House Committee on Health has launched an investigation into the management of Philippine Health Insurance Corporation (PhilHealth) funds following reports of significant delays in healthcare provider reimbursements and concerns about reserve fund utilization.</p>

<h2>Scope of Investigation</h2>
<p>Committee Chair Representative Angelina Tan announced that the inquiry will examine PhilHealth's financial management from 2022 to present, focusing on claims processing timelines, reserve fund investments, and benefit package implementation.</p>

<h2>Healthcare Provider Complaints</h2>
<p>Multiple hospital associations have submitted position papers detailing reimbursement delays averaging 180 days, significantly longer than the mandated 60-day period. Several smaller hospitals reported financial difficulties due to unpaid claims.</p>

<blockquote>"Our members are being forced to choose between maintaining quality care and financial survival. This situation is unsustainable," said Dr. Jaime Galvez Tan, representing the Private Hospital Association.</blockquote>

<h2>PhilHealth Response</h2>
<p>PhilHealth President Emmanuel Ledesma defended the agency's performance, citing improved processing rates in recent months. He attributed earlier delays to system upgrades and enhanced fraud detection measures.</p>

<h2>Reserve Fund Questions</h2>
<p>Lawmakers have also questioned the accumulation of reserve funds exceeding PHP 500 billion while members face coverage gaps. PhilHealth officials explained that reserves are maintained as required by law for long-term sustainability.</p>

<h2>Next Steps</h2>
<p>The committee has scheduled three more hearings before submitting its findings. Legislative recommendations may include amendments to the Universal Health Care Law and enhanced Congressional oversight of PhilHealth operations.</p>`,
			categorySlug: "investigations",
			tagSlugs:     []string{"house-of-representatives", "doh", "healthcare", "accountability", "exclusive"},
		},
		{
			slug:    "supreme-court-upholds-mining-regulations-environmental-protection",
			title:   "Supreme Court Upholds Stricter Mining Regulations in Landmark Environmental Decision",
			summary: "The High Tribunal rules in favor of enhanced environmental protection requirements for mining operations, affirming DENR authority.",
			content: `<p>In a landmark decision, the Supreme Court has upheld the constitutionality of Department of Environment and Natural Resources (DENR) Administrative Order No. 2023-15, which imposes stricter environmental protection requirements on mining operations throughout the Philippines.</p>

<h2>The Decision</h2>
<p>Writing for the majority, Associate Justice Marvic Leonen emphasized the State's constitutional mandate to protect the environment and the right of Filipinos to a balanced and healthful ecology. The 12-3 decision rejected challenges from mining industry groups.</p>

<blockquote>"The exploitation of natural resources must be balanced against our duty to preserve the environment for future generations. The Constitution is clear on this mandate," Justice Leonen wrote.</blockquote>

<h2>New Requirements</h2>
<p>The upheld regulations require mining companies to maintain higher financial guarantees for rehabilitation, submit comprehensive environmental impact studies, and allocate a larger percentage of revenues for affected community development.</p>

<h2>Industry Impact</h2>
<p>The Chamber of Mines of the Philippines expressed disappointment but committed to compliance. Industry analysts project that some smaller mining operations may consolidate or cease operations due to increased costs.</p>

<h2>Environmental Groups Welcome Decision</h2>
<p>Environmental advocates praised the ruling as a victory for sustainable development. The Kalikasan People's Network called for vigorous enforcement of the regulations, particularly in areas with documented environmental violations.</p>

<h2>Economic Considerations</h2>
<p>The Department of Finance acknowledged potential short-term impacts on mining revenues but expressed confidence that responsible mining can coexist with environmental protection, supporting long-term economic sustainability.</p>`,
			categorySlug: "judiciary",
			tagSlugs:     []string{"supreme-court", "denr", "mining", "climate-change", "analysis"},
		},
		{
			slug:    "infrastructure-department-unveils-five-year-connectivity-plan",
			title:   "Infrastructure Department Unveils Five-Year Philippine Connectivity Master Plan",
			summary: "A comprehensive infrastructure roadmap targets improved transportation links between major economic centers and underserved regions.",
			content: `<p>The Department of Public Works and Highways (DPWH), in coordination with the Department of Transportation (DOTr), has unveiled the Philippine Connectivity Master Plan 2025-2030, a comprehensive infrastructure roadmap aimed at dramatically improving transportation links nationwide.</p>

<h2>Key Projects</h2>
<p>The plan prioritizes completion of major expressway networks, modernization of regional airports, expansion of the Metro Manila subway system, and development of new port facilities in Mindanao and the Visayas.</p>

<h2>Budget Allocation</h2>
<p>The government has committed PHP 8.4 trillion over five years, with approximately 60% sourced from Official Development Assistance (ODA) loans and the remainder from the national budget and Public-Private Partnerships (PPP).</p>

<blockquote>"This is the most ambitious infrastructure program in Philippine history. We are building not just roads and bridges, but the foundation for inclusive economic growth," DPWH Secretary Manuel Bonoan stated.</blockquote>

<h2>Regional Development Focus</h2>
<p>Unlike previous infrastructure programs concentrated in Luzon, the new plan allocates 40% of resources to Visayas and Mindanao projects. Priority corridors include the Davao-GenSan Economic Corridor and the Central Visayas Maritime Hub.</p>

<h2>Technology Integration</h2>
<p>The plan incorporates smart infrastructure elements, including integrated traffic management systems, digital toll collection, and infrastructure monitoring using IoT sensors. The DICT will coordinate technology standards across projects.</p>

<h2>Timeline and Monitoring</h2>
<p>A dedicated Project Management Office under NEDA will track implementation progress with quarterly public reporting. The first major project completions are targeted for 2026, with full program completion by 2030.</p>`,
			categorySlug: "policy",
			tagSlugs:     []string{"dpwh", "dotr", "infrastructure", "build-better-more", "ppp", "feature"},
		},
		{
			slug:    "bsp-maintains-policy-rate-amid-inflation-concerns",
			title:   "BSP Maintains Policy Rate Amid Persistent Inflation Concerns",
			summary: "The Bangko Sentral ng Pilipinas holds interest rates steady while signaling readiness to act if price pressures persist.",
			content: `<p>The Bangko Sentral ng Pilipinas (BSP) Monetary Board has decided to maintain the overnight reverse repurchase (RRP) rate at 6.25% following its latest policy meeting, citing persistent inflation concerns despite signs of economic stabilization.</p>

<h2>Inflation Outlook</h2>
<p>BSP Governor Eli Remolona Jr. noted that while headline inflation has moderated to 4.1%, core inflation remains elevated. The central bank's inflation forecast for 2025 stands at 3.5%, within the target range of 2-4%.</p>

<blockquote>"We remain vigilant. The risks to inflation are tilted to the upside, particularly from potential supply disruptions and global oil price volatility," Governor Remolona explained during the post-meeting press conference.</blockquote>

<h2>Economic Growth Considerations</h2>
<p>The Monetary Board acknowledged the economy's resilience, with GDP growth projected at 6.0-7.0% for 2025. However, policymakers emphasized that price stability remains the primary mandate.</p>

<h2>Market Reactions</h2>
<p>The Philippine Stock Exchange Index showed modest gains following the announcement, while the peso strengthened slightly against the US dollar. Analysts had widely anticipated the hold decision.</p>

<h2>Future Guidance</h2>
<p>The BSP signaled that rate cuts remain possible in the second half of 2025 if inflation continues its downward trajectory. However, the central bank retained flexibility to adjust policy in either direction based on data.</p>

<h2>Business Sector Response</h2>
<p>Business groups expressed understanding of the BSP's cautious stance while hoping for eventual rate reductions to support investment and consumption. The Philippine Chamber of Commerce called for continued coordination between monetary and fiscal authorities.</p>`,
			categorySlug: "policy",
			tagSlugs:     []string{"bsp", "economy", "inflation", "analysis"},
		},
		{
			slug:    "dole-implements-new-wage-adjustment-framework",
			title:   "DOLE Implements New Regional Wage Adjustment Framework Nationwide",
			summary: "The Department of Labor and Employment rolls out a revised system for determining minimum wage increases across all regions.",
			content: `<p>The Department of Labor and Employment (DOLE) has implemented a new wage adjustment framework that will guide Regional Tripartite Wages and Productivity Boards (RTWPBs) in determining minimum wage increases across the country.</p>

<h2>Framework Components</h2>
<p>The new system considers a standardized basket of goods and services, regional cost of living indices, productivity metrics, and industry capacity to pay. DOLE Secretary Bienvenido Laguesma emphasized the framework's balance between worker welfare and business sustainability.</p>

<blockquote>"Workers deserve wages that allow them to live with dignity. At the same time, we must ensure businesses can continue to operate and provide employment," Secretary Laguesma stated.</blockquote>

<h2>Regional Variations</h2>
<p>The framework acknowledges significant cost-of-living differences across regions. NCR, CALABARZON, and Central Luzon are expected to see higher wage floors, while adjustments in other regions will reflect local economic conditions.</p>

<h2>Labor Group Reactions</h2>
<p>The Trade Union Congress of the Philippines (TUCP) welcomed the standardized approach but called for more frequent review cycles. The Federation of Free Workers expressed concern that the framework may not adequately address inflation impacts.</p>

<h2>Employer Perspectives</h2>
<p>The Employers Confederation of the Philippines (ECOP) appreciated the inclusion of productivity measures and capacity-to-pay considerations. Small business groups requested transition periods for wage adjustments.</p>

<h2>Implementation Timeline</h2>
<p>RTWPBs will begin applying the new framework immediately for upcoming wage petitions. DOLE projects that most regions will complete their first reviews under the new system by Q2 2025.</p>`,
			categorySlug: "policy",
			tagSlugs:     []string{"dole", "labor", "minimum-wage", "breaking-news"},
		},
		{
			slug:    "bangsamoro-parliament-passes-revenue-sharing-code",
			title:   "Bangsamoro Parliament Passes Historic Revenue Sharing Code",
			summary: "BARMM's legislative assembly approves landmark legislation governing the distribution of national wealth within the autonomous region.",
			content: `<p>The Bangsamoro Transition Authority Parliament has passed the Bangsamoro Revenue Code, a historic piece of legislation that establishes the framework for revenue generation and sharing within the Bangsamoro Autonomous Region in Muslim Mindanao (BARMM).</p>

<h2>Revenue Allocation</h2>
<p>The code specifies how the region's share of national taxes, including the 75% share of taxes from natural resources, will be distributed among BARMM's provinces, cities, and municipalities. Chief Minister Ahod Ebrahim hailed the passage as a milestone in Bangsamoro self-governance.</p>

<blockquote>"This code is the fruit of decades of struggle. It gives our people genuine control over resources within our homeland while ensuring equitable distribution," Chief Minister Ebrahim declared.</blockquote>

<h2>Local Government Shares</h2>
<p>Component LGUs will receive increased allocations under the code, with provisions for special development funds in conflict-affected and underdeveloped areas. Indigenous Peoples' communities are guaranteed specific allocations for ancestral domain development.</p>

<h2>Natural Resource Management</h2>
<p>The legislation includes provisions for sustainable natural resource management, requiring environmental impact assessments and community consent before extractive projects proceed. Revenue from such activities will fund long-term development programs.</p>

<h2>Implementation Challenges</h2>
<p>BARMM officials acknowledged challenges in building administrative capacity to implement the complex revenue sharing scheme. The World Bank and Asian Development Bank have committed technical assistance for financial management systems.</p>

<h2>Transition Outlook</h2>
<p>With parliamentary elections scheduled for 2025, the passage of the Revenue Code represents one of the final major legislative achievements of the transition period. The elected parliament will have authority to amend the code based on implementation experience.</p>`,
			categorySlug: "local-government",
			tagSlugs:     []string{"barmm", "bangsamoro", "peace-process", "feature"},
		},
		{
			slug:    "commission-on-human-rights-releases-ejk-investigation-findings",
			title:   "Commission on Human Rights Releases Comprehensive EJK Investigation Findings",
			summary: "CHR presents multi-year investigation results on extrajudicial killings, recommending prosecutions and institutional reforms.",
			content: `<p>The Commission on Human Rights (CHR) has released its comprehensive report on extrajudicial killings (EJKs) documented during the anti-drug campaign, presenting findings from over 30,000 cases investigated since 2016.</p>

<h2>Key Findings</h2>
<p>The report documents patterns of human rights violations, identifies chains of command in implicated operations, and provides evidence packages for potential prosecutions. CHR Chairperson Richard Palpal-latoc emphasized the commission's commitment to accountability.</p>

<blockquote>"Justice delayed is justice denied. These findings represent the voices of thousands of families seeking accountability. The evidence we present meets prosecutorial standards," Chairperson Palpal-latoc stated.</blockquote>

<h2>Recommendations</h2>
<p>The CHR recommends criminal prosecution of identified perpetrators, compensation programs for victims' families, and institutional reforms within the Philippine National Police. The report also calls for enhanced witness protection mechanisms.</p>

<h2>Department of Justice Response</h2>
<p>DOJ Secretary Jesus Crispin Remulla confirmed receipt of the report and pledged thorough review of cases meeting evidentiary standards. The department has created a special task force to handle EJK-related prosecutions.</p>

<h2>International Implications</h2>
<p>The report comes as the International Criminal Court continues its investigation into alleged crimes against humanity in the Philippines. Government officials maintain that domestic mechanisms are adequate for addressing violations.</p>

<h2>Civil Society Reactions</h2>
<p>Human rights organizations praised the report's thoroughness while urging swift government action. Families of victims expressed hope that the findings would finally lead to justice for their loved ones.</p>`,
			categorySlug: "investigations",
			tagSlugs:     []string{"chr", "human-rights", "ejk", "pnp", "doj", "icc", "exclusive"},
		},
		{
			slug:    "deped-launches-basic-education-curriculum-reform",
			title:   "DepEd Launches Comprehensive Basic Education Curriculum Reform",
			summary: "The Department of Education unveils MATATAG curriculum focusing on essential competencies and reduced learning load.",
			content: `<p>The Department of Education (DepEd) has officially launched the MATATAG curriculum, a comprehensive reform of basic education that aims to address learning gaps exacerbated by the pandemic while preparing Filipino students for 21st-century challenges.</p>

<h2>Curriculum Changes</h2>
<p>The reformed curriculum reduces the number of subjects per grade level while deepening focus on essential competencies. Filipino and English literacy, mathematics, and science are prioritized, with emphasis on practical application over rote memorization.</p>

<h2>Implementation Approach</h2>
<p>DepEd Secretary Sonny Angara outlined a phased implementation beginning with Grades 1, 4, and 7 in School Year 2024-2025. Complete rollout across all grade levels will conclude by School Year 2027-2028.</p>

<blockquote>"We listened to teachers, parents, and students. This curriculum is designed for meaningful learning, not just coverage of topics. Quality over quantity is our principle," Secretary Angara explained.</blockquote>

<h2>Teacher Training</h2>
<p>A massive teacher training program is underway, with over 800,000 public school teachers scheduled for curriculum orientation. International partners including USAID and JICA are supporting training material development.</p>

<h2>Learning Resources</h2>
<p>New textbooks and learning materials aligned with MATATAG are being developed and distributed. Digital resources will complement physical materials, addressing the varying technology access across schools.</p>

<h2>Assessment Reform</h2>
<p>The new curriculum includes reformed assessment approaches, moving away from purely test-based evaluation toward competency demonstrations and portfolio assessments. National Achievement Tests will be redesigned to align with curriculum changes.</p>`,
			categorySlug: "policy",
			tagSlugs:     []string{"deped", "education", "feature"},
		},
	}

	for _, article := range articles {
		// Check if article already exists
		var exists bool
		err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM articles WHERE slug = $1)`, article.slug).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check article existence: %w", err)
		}
		if exists {
			fmt.Printf("  - Article '%s' already exists, skipping\n", article.title[:50]+"...")
			continue
		}

		categoryID := categoryIDs[article.categorySlug]

		// Insert article
		var articleID string
		err = conn.QueryRow(ctx, `
			INSERT INTO articles (slug, title, summary, content, author_id, category_id, status)
			VALUES ($1, $2, $3, $4, $5, $6, 'draft')
			RETURNING id
		`, article.slug, article.title, article.summary, article.content, authorID, categoryID).Scan(&articleID)
		if err != nil {
			return fmt.Errorf("failed to insert article %s: %w", article.slug, err)
		}

		// Insert article tags
		for _, tagSlug := range article.tagSlugs {
			tagID, ok := tagIDs[tagSlug]
			if !ok {
				continue
			}
			_, err = conn.Exec(ctx, `
				INSERT INTO article_tags (article_id, tag_id)
				VALUES ($1, $2)
				ON CONFLICT DO NOTHING
			`, articleID, tagID)
			if err != nil {
				return fmt.Errorf("failed to insert article tag: %w", err)
			}
		}

		fmt.Printf("  - Article '%s' seeded\n", article.title[:50]+"...")
	}

	return nil
}
